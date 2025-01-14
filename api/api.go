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
	Secret  string
}

type Item struct {
	Project     int
	Name        string
	Price       int
	Picture     string
	Description string
	PrevPrice   int
	ID          int
	Category    string
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
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
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
		if err := rows.Scan(&project.ID, &project.Name, &project.IPAddr, &project.Balance, &project.Owner, &project.Secret); err != nil {
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
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
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

	_, err = db.Exec("INSERT INTO projects (name, ipaddr, balance, owner, secret) VALUES (?, ?, ?, ?, ?)", name, "localhost", 0.00, steamid, auth.GenerateRandomSessionToken())
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
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
		return
	}

	steamid, exists := auth.ValidateUserSession(cookie.Value)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusUnauthorized)
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
	category := req.PostFormValue("category")

	if project == "" || name == "" || price == "" || image == "" || description == "" || previous_price == "" {
		http.Error(w, "Something is null :(", http.StatusBadRequest)
		return
	}

	if category == "" {
		category = "Разное"
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

	_, err = db.Exec("INSERT INTO items (project, name, price, picture, description, previous_price, category) VALUES (?, ?, ?, ?, ?, ?, ?)", newProject, name, newPrice, image, description, newPrevPrice, category)
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
		http.Error(w, "failed to convert project id to an integer", http.StatusInternalServerError)
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
		if err := rows.Scan(&item.Project, &item.Name, &item.Price, &item.Picture, &item.Description, &item.PrevPrice, &item.ID, &item.Category); err != nil {
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
	var projectID int
	var ownerSteamID string
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusNotAcceptable)
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

	item := req.PostFormValue("itemid")
	newItem, err := strconv.Atoi(item)
	if err != nil {
		http.Error(w, "Couldn't convert item id to an integer", http.StatusInternalServerError)
		return
	}

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

func EditProject(w http.ResponseWriter, req *http.Request) {
	var projectExists bool
	project, err := strconv.Atoi(req.PostFormValue("projectid"))
	if err != nil {
		http.Error(w, "Couldn't convert project ID to an integer", http.StatusInternalServerError)
		return
	}
	newName := req.PostFormValue("newname")

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusNotAcceptable)
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

	_, err = db.Exec("UPDATE projects SET name = ? WHERE id = ?", newName, project)
	if err != nil {
		http.Error(w, "something's up with our backend :/", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}

func EditItem(w http.ResponseWriter, req *http.Request) {
	var ownerSteamID string
	var projectID int
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusNotAcceptable)
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

	item := req.PostFormValue("itemid")
	newItem, err := strconv.Atoi(item)
	if err != nil {
		http.Error(w, "Couldn't convert item id to an integer", http.StatusInternalServerError)
		return
	}

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

	/* newName := req.PostFormValue("newname")
	newPrice, err := strconv.Atoi(req.PostFormValue("newprice"))
	if err != nil {
		http.Error(w, "couldn't convert newPrice to an integer", http.StatusInternalServerError)
		return
	}
	newPicture := req.PostFormValue("newpicture")
	newDescription := req.PostFormValue("newdescription")
	newPrevPrice, err := strconv.Atoi(req.PostFormValue("newprevprice"))
	if err != nil {
		http.Error(w, "couldn't convert newPrevPrice to an integer", http.StatusInternalServerError)
		return
	}
	newCategory := req.PostFormValue("newcategory")

	_, err = db.Exec("UPDATE items SET name = ?, price = ?, picture = ?, description = ?, previous_price = ?, category = ? WHERE id = ?", newName, newPrice, newPicture, newDescription, newPrevPrice, newCategory, item)
	if err != nil {
		http.Error(w, "something's up with the backend", http.StatusInternalServerError)
		return
	} */

	changes := map[string]string{
		"name":        req.PostFormValue("newname"),
		"picture":     req.PostFormValue("newpicture"),
		"description": req.PostFormValue("newdescription"),
		"category":    req.PostFormValue("newcategory"),
	}
	var price, prevprice int

	if req.PostFormValue("newprice") != "" {
		price, err = strconv.Atoi(req.PostFormValue("newprice"))
		if err != nil {
			http.Error(w, "invalid price", http.StatusInternalServerError)
			return
		}
	}

	if req.PostFormValue("newprevprice") != "" {
		prevprice, err = strconv.Atoi(req.PostFormValue("newprevprice"))
		if err != nil {
			http.Error(w, "invalid prevprice", http.StatusInternalServerError)
			return
		}
	}

	_, err = db.Exec("UPDATE items SET price = ?, previous_price = ? WHERE id = ?", price, prevprice, item)
	if err != nil {
		http.Error(w, "something's up with the database", http.StatusInternalServerError)
	}

	for i, v := range changes {
		if v == "" {
			var new string
			err = db.QueryRow("SELECT "+i+" FROM items WHERE id = ?", item).Scan(&new)
			if err != nil {
				http.Error(w, "something's up with the backend", http.StatusInternalServerError)
				return
			}
			changes[i] = new
		}
	}

	for i, v := range changes {
		_, err = db.Exec("UPDATE items SET "+i+" = ? WHERE id = ?", v, item)
		if err != nil {
			http.Error(w, "something's up with the backend", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}
