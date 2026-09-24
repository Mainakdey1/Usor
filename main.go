package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from go")

}

func main() {
	http.HandleFunc("/", hello)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
