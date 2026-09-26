package main

import (
	"Usor/utilities"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from go")

}

type RegisterUser struct {
	Username string `json:"user"`
	Password string `json:"password"`
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func register(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var user RegisterUser

		//Decoder takes body from request json and decodes to user
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		fmt.Println("User: ", user.Username)

		passwordHash, err := utilities.HashPassword(user.Password)

		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			r.Context(),
			`INSERT INTO users (id, username, password_hash)
			VALUES ($1, $2, $3)`,
			uuid.New(),
			user.Username,
			passwordHash,
		)

		if err != nil {
			log.Println(err)
			http.Error(w, "failed to write to db", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"message": "user registered!",
		})
	}
}

func main() {
	port := os.Getenv("PORT")

	dbUrl := os.Getenv("DATABASE_URL")

	db, err := pgxpool.New(context.Background(), dbUrl)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", hello)
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /register", register(db))

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if port == "" {
		port = "8000"
	}
	println("Server running on :", port)
	err = http.ListenAndServe(":"+port, mux)

	if err != nil {
		panic(err)
	}
}
