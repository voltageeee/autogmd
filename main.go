package main

import (
	"autogmd/api"
	"autogmd/auth"
	"net/http"
)

func main() {
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/steam/callback", auth.SteamCallbackHandler)
	http.HandleFunc("/logout", auth.Logout)
	http.HandleFunc("/protected", auth.ProtectedHandler)
	http.HandleFunc("/api/projects/create", api.NewProject)
	http.HandleFunc("/api/projects/get", api.GetProjects)
	http.HandleFunc("/api/projects/delete", api.DeleteProject)
	http.ListenAndServe(":8080", nil)
}
