package api

import (
	"autogmd/auth"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func GetProjects(w http.ResponseWriter, req *http.Request) {
	var projects []Project

	steamid, exists := auth.ValidateUserSession(req, w)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
		return
	}

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
	steamid, exists := auth.ValidateUserSession(req, w)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
		return
	}

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
	steamid, exists := auth.ValidateUserSession(req, w)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
		return
	}

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

	steamid, exists := auth.ValidateUserSession(req, w)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

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

	steamid, exists := auth.ValidateUserSession(req, w)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
		return
	}

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
