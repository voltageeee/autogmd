package api

import (
	"autogmd/auth"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// структура нашего проекта
type Project struct {
	ID      int
	Name    string
	IPAddr  string
	Balance float32
	Owner   string
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

func GetProjects(w http.ResponseWriter, req *http.Request) {
	var projects []Project

	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := req.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, "http://localhost:8080/login", http.StatusTemporaryRedirect)
		return
	}

	rows, err := db.Query("SELECT * FROM projects WHERE owner = ?", steamid)
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Name, &project.IPAddr, &project.Balance, &project.Owner); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Failed to read data", http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error during rows iteration:", err)
		http.Error(w, "Failed to read data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func NewProject(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := req.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, "http://localhost:8080/login", http.StatusTemporaryRedirect)
		return
	}

	name := req.PostFormValue("projectname")

	var projectExists bool

	err = db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE name = ? AND owner = ?", name, steamid).Scan(&projectExists)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "our backend is fucked, try again later.", http.StatusInternalServerError)
		return
	}

	if projectExists {
		http.Error(w, "Project already exists", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO projects (name, ipaddr, balance, owner) VALUES (?, ?, ?, ?)", name, "localhost", 0.00, steamid)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "our backend is fucked, try again later.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}

func DeleteProject(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := req.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, "http://localhost:8080/login", http.StatusTemporaryRedirect)
		return
	}

	name := req.PostFormValue("projectname")

	var projectExists bool

	err = db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE name = ? AND owner = ?", name, steamid).Scan(&projectExists)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "our backend is fucked, try again later.", http.StatusInternalServerError)
		return
	}

	if !projectExists {
		http.Error(w, "Project doesn't exist", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM projects WHERE name = ? AND owner = ?", name, steamid)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "our backend is fucked, try again later.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}
