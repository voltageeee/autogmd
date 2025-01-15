package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

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

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	_, err = os.Create("logs.txt")
	if err != nil {
		log.Fatal("failed to create log file!")
	}
}

func handleError(err error, w http.ResponseWriter, msg string) {
	if err == nil {
		return
	}

	logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()

	currentTime := time.Now().Format(time.RFC3339)

	logMessage := fmt.Sprintf("error: %v | msg sent: %s | timestamp: %s\n", err, msg, currentTime)
	logFile.Write([]byte(logMessage))

	log.Println(logMessage)

	http.Error(w, msg, http.StatusInternalServerError)
}

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
