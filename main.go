package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from go")

}

type RegisterUser struct {
	UserID   string `json:"user"`
	Password string `json:"password"`
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func user_register(w http.ResponseWriter, r *http.Request) {
	dBUrl := os.Getenv("DATABASE_URL")
	w.Header().Set("Content-Type", "application/json")

	var user RegisterUser

	//Decoder takes body from request json and decodes to user
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	fmt.Println("User: ", user.UserID)
	fmt.Println("Pass: ", user.Password)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "user registered!",
	})
}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/health", health)
	http.HandleFunc("/register", user_register)
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
