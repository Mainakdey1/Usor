package main

import (
	"Usor/utilities"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
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
		uuidToken := uuid.New()
		_, err = db.Exec(
			r.Context(),
			`INSERT INTO users (id, username, password_hash)
			VALUES ($1, $2, $3)`,
			uuidToken,
			user.Username,
			passwordHash,
		)

		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				http.Error(w, "Username already exists",
					http.StatusConflict)
			} else {
				log.Printf("Database err: %v", err)
				http.Error(w, "Internal Server Error",
					http.StatusInternalServerError)
			}
			return
		}
		token, err := utilities.GenerateToken()
		if err != nil {
			panic(err)
		}
		tokenHash := utilities.HashToken(token)
		_, err = db.Exec(
			r.Context(),
			`INSERT INTO api_tokens (user_id, token_hash)
			VALUES ($1, $2)`,
			uuidToken,
			tokenHash,
		)
		if err != nil {
			log.Printf("Database err: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"message": "user registered!",
		})
	}
}

func main() {
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading env")
		}
	}
	port := os.Getenv("PORT")

	dbUrl := os.Getenv("DATABASE_URL")

	db, err := pgxpool.New(context.Background(), dbUrl)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", hello)
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /register", register(db))

	handler := corsMiddleware(mux)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if port == "" {
		port = "8000"
	}
	println("Server running on :", port)
	err = http.ListenAndServe(":"+port, handler)

	if err != nil {
		panic(err)
	}
}
