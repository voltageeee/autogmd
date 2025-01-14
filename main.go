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
	http.HandleFunc("/api/items/create", api.NewItem)
	http.HandleFunc("/api/items/get", api.GetItems)
	http.HandleFunc("/aöi/items/delete", api.DeleteItem)
	http.ListenAndServe(":8080", nil)
}
