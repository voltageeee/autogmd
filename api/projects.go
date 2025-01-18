package api

import (
	"autogmd/auth"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Code struct {
	Code      string    `json:"code"`
	Timestamp time.Time `json:"timestamp"`
	Expires   time.Time `json:"expires"`
}

var codes = map[Code]int{}

func GetProjects(w http.ResponseWriter, req *http.Request) {
	var projects []Project

	steamid := req.Context().Value(auth.SteamIDKey).(string)

	rows, err := db.Query(`
		SELECT p.id, p.name, p.ipaddr, p.balance, p.owner, p.secret
		FROM projects p
		WHERE p.owner = ? OR p.id IN (SELECT project_id FROM coowners WHERE coowner_id = ?)
		`, steamid, steamid)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Name, &project.IPAddr, &project.Balance, &project.Owner, &project.Secret); err != nil {
			handleError(err, w, "We can't process your request right now.")
			return
		}

		_, isCoowner, err := auth.VerifyOwnership(steamid, project.ID)
		if err != nil {
			log.Println("Ownership verification failed:", err)
			continue
		}

		if isCoowner {
			project.Secret = ""
		}

		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}
}

func NewProject(w http.ResponseWriter, req *http.Request) {
	steamid := req.Context().Value(auth.SteamIDKey).(string)

	name := req.PostFormValue("projectname")

	var projectExists bool

	err := db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE name = ? AND owner = ?", name, steamid).Scan(&projectExists)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	if projectExists {
		http.Error(w, "Project already exists", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO projects (name, ipaddr, balance, owner, secret) VALUES (?, ?, ?, ?, ?)", name, "localhost", 0.00, steamid, auth.GenerateRandomSessionToken())
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteProject(w http.ResponseWriter, req *http.Request) {
	steamid := req.Context().Value(auth.SteamIDKey).(string)

	project := req.PostFormValue("projectid")
	projectid, err := strconv.Atoi(project)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
	}

	isOwner, _, err := auth.VerifyOwnership(steamid, projectid)
	if err != nil {
		handleError(err, w, "Failed to verify ownership.")
		return
	}
	if !isOwner {
		http.Error(w, "You aren't the owner", http.StatusUnauthorized)
		return
	}

	_, err = db.Exec("DELETE FROM projects WHERE id = ? AND owner = ?", projectid, steamid)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func EditProject(w http.ResponseWriter, req *http.Request) {
	project, err := strconv.Atoi(req.PostFormValue("projectid"))
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}
	newName := req.PostFormValue("newname")

	steamid := req.Context().Value(auth.SteamIDKey).(string)

	isOwner, _, err := auth.VerifyOwnership(steamid, project)
	if err != nil {
		handleError(err, w, "Failed to verify ownership.")
		return
	}
	if !isOwner {
		http.Error(w, "You aren't the owner", http.StatusUnauthorized)
		return
	}

	_, err = db.Exec("UPDATE projects SET name = ? WHERE id = ?", newName, project)
	if err != nil {
		http.Error(w, "something's up with our backend :/", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func AddCoOwner(w http.ResponseWriter, req *http.Request) {
	coownerId := req.PostFormValue("coownerid")

	steamid := req.Context().Value(auth.SteamIDKey).(string)
	var exists bool

	id, err := strconv.Atoi(req.PostFormValue("projectid"))
	if err != nil {
		handleError(err, w, "Project ID must be a positive number.")
		return
	}

	if id <= 0 {
		http.Error(w, "Project ID must be positive", http.StatusBadRequest)
		return
	}

	isOwner, _, err := auth.VerifyOwnership(steamid, id)
	if err != nil {
		handleError(err, w, "Failed to verify project ownership.")
		return
	}

	if !isOwner {
		http.Error(w, "You aren't the project's owner", http.StatusUnauthorized)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) > 0 FROM users WHERE id = ?", coownerId).Scan(&exists)
	if err != nil {
		handleError(err, w, "Sorry, we're unable to process your request right now")
	}

	if !exists {
		http.Error(w, "Coowner must be registered as a valid user before proceeding.", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO coowners (project_id, coowner_id) VALUES (?, ?)", id, coownerId)
	if err != nil {
		handleError(err, w, "Sorry, we're unable to process your request right now.")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Register(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PostFormValue("projectid"))
	if err != nil {
		handleError(err, w, "seems like you're trying to pass a non-int value")
		return
	}

	if id <= 0 {
		http.Error(w, "projectid must be a positive integer", http.StatusBadRequest)
		return
	}

	var exists bool

	err = db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE id = ?", id).Scan(&exists)
	if err != nil {
		http.Error(w, "project must exist", http.StatusBadRequest)
		return
	}

	code := Code{
		Code:      auth.GenerateRandomSessionToken(),
		Timestamp: time.Now(),
		Expires:   time.Now().Add(time.Second * 180),
	}

	codes[code] = id

	w.WriteHeader(http.StatusOK)
}

func Confirm(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PostFormValue("projectid"))
	if err != nil {
		handleError(err, w, "seems like you're trying to pass a non-int value")
		return
	}
	codeKey := req.PostFormValue("confirmationcode")
	ipaddr := req.PostFormValue("ipaddr")

	if id <= 0 {
		http.Error(w, "projectid must be a positive integer", http.StatusBadRequest)
		return
	}

	var exists bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE id = ?", id).Scan(&exists)
	if err != nil {
		handleError(err, w, "something's wrong on our side, sorry")
		return
	}

	for code, codeId := range codes {
		if code.Code == codeKey {
			if code.Expires.Before(time.Now()) {
				delete(codes, code)
				http.Error(w, "Confirmation code has expired", http.StatusBadRequest)
				return
			}

			if id != codeId {
				http.Error(w, "Project ID does not match the confirmation code", http.StatusBadRequest)
				return
			}

			_, err = db.Exec("UPDATE projects SET ipaddr = ? WHERE id = ?", ipaddr, id)
			if err != nil {
				handleError(err, w, "Database update failed")
				return
			}

			delete(codes, code)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Invalid confirmation code", http.StatusBadRequest)
}

func GetConfirmationCodes(w http.ResponseWriter, req *http.Request) {
	steamid := req.Context().Value(auth.SteamIDKey).(string)

	id, err := strconv.Atoi(req.PostFormValue("projectid"))
	if err != nil {
		handleError(err, w, "seems like you're trying to pass a non-integer value")
		return
	}

	isOwner, _, err := auth.VerifyOwnership(steamid, id)
	if !isOwner {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You are not the project owner"))
	}
	if err != nil {
		handleError(err, w, "something's wrong on our side")
	}

	codesToSend := []Code{}

	for i, v := range codes {
		if v == id {
			codesToSend = append(codesToSend, i)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(codesToSend); err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}
}
