package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type User struct {
	Username   string `json:"username"`
	SteamID    string `json:"steamid"`
	ProfilePic string `json:"profile_pic"`
}

var db *sql.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}
	db, err = sql.Open("mysql", os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
}

func upsertUser(id, username, sessionToken string) error {
	var existingToken string
	err := db.QueryRow("SELECT session_token FROM users WHERE id = ?", id).Scan(&existingToken)
	if err == sql.ErrNoRows {
		_, err := db.Exec("INSERT INTO users (id, username, session_token) VALUES (?, ?, ?)", id, username, sessionToken)
		if err != nil {
			return fmt.Errorf("error inserting new user: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error querying user: %v", err)
	} else {
		if existingToken != sessionToken {
			_, err := db.Exec("UPDATE users SET session_token = ? WHERE id = ?", sessionToken, id)
			if err != nil {
				return fmt.Errorf("error updating session token: %v", err)
			}
		}
	}

	return nil
}

func ValidateUserSession(session string) (string, bool) {
	var steamID string
	err := db.QueryRow("SELECT id FROM users WHERE session_token = ?", session).Scan(&steamID)
	if err == sql.ErrNoRows {
		return "", false
	}

	return steamID, true
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	steamOpenIDURL := "https://steamcommunity.com/openid/login"
	params := url.Values{
		"openid.ns":         {"http://specs.openid.net/auth/2.0"},
		"openid.mode":       {"checkid_setup"},
		"openid.return_to":  {fmt.Sprintf("%s/steam/callback", os.Getenv("HOSTNAME"))},
		"openid.realm":      {os.Getenv("HOSTNAME")},
		"openid.identity":   {"http://specs.openid.net/auth/2.0/identifier_select"},
		"openid.claimed_id": {"http://specs.openid.net/auth/2.0/identifier_select"},
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, steamOpenIDURL+"?"+params.Encode(), http.StatusFound)
		return
	}

	_, exists := ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, r, steamOpenIDURL+"?"+params.Encode(), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/protected", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
}

func SteamCallbackHandler(w http.ResponseWriter, r *http.Request) {
	steamID := extractSteamID(r)
	if steamID == "" {
		http.Error(w, "Failed to verify SteamID", http.StatusForbidden)
		return
	}

	sessionToken := GenerateRandomSessionToken()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
	})

	user, err := fetchSteamUserInfo(steamID)
	if err != nil {
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
	}

	if err := upsertUser(steamID, user.Username, sessionToken); err != nil {
		log.Fatal("failed to upsert user: ", err)
	}

	http.Redirect(w, r, fmt.Sprintf("%s/protected", os.Getenv("HOSTNAME")), http.StatusFound)
}

// хуета чтоб убедиться в том, что мы залогинились. прост возвращает ник, стимайди и ссылку на аватарку в джсон формате. можно юзать для получения этих данных наврен :/
func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	steamid, exists := ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, r, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	user, err := fetchSteamUserInfo(steamid)
	if err != nil {
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func extractSteamID(r *http.Request) string {
	claimedID := r.URL.Query().Get("openid.claimed_id")
	if claimedID == "" {
		return ""
	}

	steamIDParts := strings.Split(claimedID, "/")
	return steamIDParts[len(steamIDParts)-1]
}

func fetchSteamUserInfo(steamID string) (*User, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}
	apiURL := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s", os.Getenv("STEAM_API_KEY"), steamID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	type SteamResponse struct {
		Response struct {
			Players []struct {
				SteamID   string `json:"steamid"`
				Username  string `json:"personaname"`
				AvatarURL string `json:"avatar"`
			} `json:"players"`
		} `json:"response"`
	}

	var steamResponse SteamResponse
	if err := json.Unmarshal(body, &steamResponse); err != nil {
		return nil, err
	}

	if len(steamResponse.Response.Players) == 0 {
		return nil, fmt.Errorf("no user found for steamid %s", steamID)
	}

	player := steamResponse.Response.Players[0]
	return &User{
		Username:   player.Username,
		SteamID:    player.SteamID,
		ProfilePic: player.AvatarURL,
	}, nil
}

func GenerateRandomSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func Logout(w http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("session_token")
	if err != nil {
		http.Error(w, "Oops, something's fishy! You don't seem to have your cookies in order. That must mean that you are already logged out. What a boomer, er? Hi from voltage btw", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE users SET session_token = ? WHERE session_token = ?", "", token.Value)
	if err != nil {
		http.Error(w, "I failed as a programmer...", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
}
