package api

import (
	"autogmd/auth"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func NewItem(w http.ResponseWriter, req *http.Request) {
	steamid, exists := auth.ValidateUserSession(req, w)
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
		handleError(err, w, "We can't process your request right now.")
		return
	}

	newPrevPrice, err := strconv.Atoi(previous_price)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	newProject, err := strconv.Atoi(project)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	isOwner, _, err := auth.VerifyOwnership(steamid, newProject)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}
	if !isOwner {
		http.Error(w, "You aren't the owner", http.StatusUnauthorized)
		return
	}

	_, err = db.Exec("INSERT INTO items (project, name, price, picture, description, previous_price, category) VALUES (?, ?, ?, ?, ?, ?, ?)", newProject, name, newPrice, image, description, newPrevPrice, category)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}

func GetItems(w http.ResponseWriter, req *http.Request) {
	var projectExists bool
	var items []Item

	project := req.PostFormValue("project")
	newProject, err := strconv.Atoi(project)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	err = db.QueryRow("SELECT COUNT(*) > 0 FROM projects WHERE id = ?", newProject).Scan(&projectExists)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	if !projectExists {
		http.Error(w, "Project doesn't exist", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT * FROM items WHERE project = ?", newProject)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Project, &item.Name, &item.Price, &item.Picture, &item.Description, &item.PrevPrice, &item.ID, &item.Category); err != nil {
			handleError(err, w, "We can't process your request right now.")
			return
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func DeleteItem(w http.ResponseWriter, req *http.Request) {
	var projectID int

	steamid, exists := auth.ValidateUserSession(req, w)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	item := req.PostFormValue("itemid")
	newItem, err := strconv.Atoi(item)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	err = db.QueryRow("SELECT project FROM items WHERE id = ?", newItem).Scan(&projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Couldn't find the item", http.StatusBadRequest)
		return
	}

	isOwner, _, err := auth.VerifyOwnership(steamid, projectID)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}
	if !isOwner {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	_, err = db.Exec("DELETE FROM items WHERE id = ?", newItem)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}

func EditItem(w http.ResponseWriter, req *http.Request) {
	var projectID int

	steamid, exists := auth.ValidateUserSession(req, w)
	if !exists {
		http.Redirect(w, req, fmt.Sprintf("%s/login", os.Getenv("HOSTNAME")), http.StatusTemporaryRedirect)
		return
	}

	isOwner, _, err := auth.VerifyOwnership(steamid, projectID)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	if !isOwner {
		http.Error(w, "You are not the project owner", http.StatusUnauthorized)
		return
	}

	item := req.PostFormValue("itemid")
	newItem, err := strconv.Atoi(item)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
		return
	}

	err = db.QueryRow("SELECT project FROM items WHERE id = ?", newItem).Scan(&projectID)
	if err == sql.ErrNoRows {
		handleError(err, w, "We can't process your request right now.")
		return
	}

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
			handleError(err, w, "We can't process your request right now.")
			return
		}
	}

	if req.PostFormValue("newprevprice") != "" {
		prevprice, err = strconv.Atoi(req.PostFormValue("newprevprice"))
		if err != nil {
			handleError(err, w, "We can't process your request right now.")
			return
		}
	}

	_, err = db.Exec("UPDATE items SET price = ?, previous_price = ? WHERE id = ?", price, prevprice, item)
	if err != nil {
		handleError(err, w, "We can't process your request right now.")
	}

	for i, v := range changes {
		if v == "" {
			var new string
			err = db.QueryRow(fmt.Sprintf("UPDATE items SET %s = ? WHERE id = ?", i), item).Scan(&new)
			if err != nil {
				handleError(err, w, "We can't process your request right now.")
				return
			}
			changes[i] = new
		}
	}

	for i, v := range changes {
		_, err = db.Exec(fmt.Sprintf("UPDATE items SET %s = ? WHERE id = ?", i), v, item)
		if err != nil {
			handleError(err, w, "We can't process your request right now.")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("gotcha!"))
}
