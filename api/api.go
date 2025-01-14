package api

import (
	"autogmd/auth"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Project struct {
	ID      int
	Name    string
	IPAddr  string
	Balance float32
	Owner   string
}

type Item struct {
	Project     int
	Name        string
	Price       int
	Picture     string
	Description string
	PrevPrice   int
	ID          int
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
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
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
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
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
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
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

func NewItem(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := req.Cookie("session_token")
	if err != nil {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	project := req.PostFormValue("project")
	name := req.PostFormValue("name")
	price := req.PostFormValue("price")
	image := req.PostFormValue("image")
	description := req.PostFormValue("description")
	previous_price := req.PostFormValue("previousprice")

	if project == "" || name == "" || price == "" || image == "" || description == "" || previous_price == "" {
		http.Error(w, "Something is null :(", http.StatusBadRequest)
		return
	}

	newPrice, err := strconv.Atoi(price)
	if err != nil {
		http.Error(w, "couldn't convert price to an integer, seems like ur pretty fucking dumb", http.StatusInternalServerError)
		return
	}

	newPrevPrice, err := strconv.Atoi(previous_price)
	if err != nil {
		http.Error(w, "couldn't convert price to an integer, seemls like ur pretty fucking dumb", http.StatusInternalServerError)
		return
	}

	newProject, err := strconv.Atoi(project)
	if err != nil {
		http.Error(w, "couldn't conver project id to an integer, seems like ur pretty fucking dumb", http.StatusInternalServerError)
		return
	}

	var projectExists bool

	err = db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE id = ? AND owner = ?", project, steamid).Scan(&projectExists)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "our backend is fucked, try again later.", http.StatusInternalServerError)
		return
	}

	if !projectExists {
		http.Error(w, "Project doesn't exist or you aren't the owner", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO items (project, name, price, picture, description, previous_price) VALUES (?, ?, ?, ?, ?, ?)", newProject, name, newPrice, image, description, newPrevPrice)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "our backend is fucked, try again later.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}

func GetItems(w http.ResponseWriter, req *http.Request) {
	var projectExists bool
	var items []Item
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	project := req.PostFormValue("project")
	newProject, err := strconv.Atoi(project)
	if err != nil {
		http.Error(w, "failed to convert shit to an integer", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE id = ?", newProject).Scan(&projectExists)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "our backend is fucked, try again later.", http.StatusInternalServerError)
		return
	}

	if !projectExists {
		http.Error(w, "Project doesn't exist", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT * FROM items WHERE project = ?", newProject)
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Project, &item.Name, &item.Price, &item.Picture, &item.Description, &item.PrevPrice, &item.ID); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Failed to read data", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error during rows iteration:", err)
		http.Error(w, "Failed to read data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func DeleteItem(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusNotAcceptable)
		return
	}

	cookie, err := req.Cookie("session_token")
	if err != nil {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	item := req.PostFormValue("itemid")
	newItem, err := strconv.Atoi(item)
	if err != nil {
		http.Error(w, "Couldn't convert item id to an integer", http.StatusInternalServerError)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	var projectID int
	var ownerSteamID string

	err = db.QueryRow("SELECT project FROM items WHERE id = ?", newItem).Scan(&projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Couldn't find the item", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("SELECT owner FROM projects WHERE id = ?", projectID).Scan(&ownerSteamID)
	if err == sql.ErrNoRows {
		http.Error(w, "oops, we fucked something up pretty bad...", http.StatusInternalServerError)
		return
	}

	if steamid != ownerSteamID {
		http.Error(w, "This is not your project. Fuck off!", http.StatusUnauthorized)
		return
	}

	_, err = db.Exec("DELETE FROM items WHERE id = ?", newItem)
	if err != nil {
		http.Error(w, "Our backend is fucked, try again later", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}