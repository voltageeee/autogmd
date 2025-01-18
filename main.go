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
	r.HandleFunc("/logout", auth.ValidateUserSession(auth.Logout)).Methods("GET")
	r.HandleFunc("/protected", auth.ValidateUserSession(auth.ProtectedHandler)).Methods("GET")

	r.HandleFunc("/api/projects/create", auth.ValidateUserSession(api.NewProject)).Methods("POST")
	r.HandleFunc("/api/projects/get", auth.ValidateUserSession(api.GetProjects)).Methods("GET")
	r.HandleFunc("/api/projects/delete", auth.ValidateUserSession(api.DeleteProject)).Methods("POST")
	r.HandleFunc("/api/projects/edit", auth.ValidateUserSession(api.EditProject)).Methods("POST")
	r.HandleFunc("/api/projects/register", api.Register).Methods("POST")
	r.HandleFunc("/api/projects/confirm", api.Confirm).Methods("POST")
	r.HandleFunc("/api/projects/getcodes", auth.ValidateUserSession(api.GetConfirmationCodes)).Methods("GET")

	r.HandleFunc("/api/items/create", auth.ValidateUserSession(api.NewItem)).Methods("POST")
	r.HandleFunc("/api/items/get", api.GetItems).Methods("GET")
	r.HandleFunc("/api/items/delete", auth.ValidateUserSession(api.DeleteItem)).Methods("POST")
	r.HandleFunc("/api/items/edit", auth.ValidateUserSession(api.EditItem)).Methods("POST")

	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}
