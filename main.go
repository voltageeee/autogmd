package main

import (
	"autogmd/api"
	"autogmd/auth"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/login", auth.LoginHandler).Methods("GET")
	r.HandleFunc("/steam/callback", auth.SteamCallbackHandler).Methods("GET")
	r.HandleFunc("/logout", auth.Logout).Methods("GET")
	r.HandleFunc("/protected", auth.ProtectedHandler).Methods("GET")

	r.HandleFunc("/api/projects/create", api.NewProject).Methods("POST")
	r.HandleFunc("/api/projects/get", api.GetProjects).Methods("GET")
	r.HandleFunc("/api/projects/delete", api.DeleteProject).Methods("POST")
	r.HandleFunc("/api/projects/edit", api.EditProject).Methods("POST")

	r.HandleFunc("/api/items/create", api.NewItem).Methods("POST")
	r.HandleFunc("/api/items/get", api.GetItems).Methods("GET")
	r.HandleFunc("/api/items/delete", api.DeleteItem).Methods("POST")
	r.HandleFunc("/api/items/edit", api.EditItem).Methods("POST")

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}
