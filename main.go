package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func echo(w http.ResponseWriter, r *http.Request) {
	values := mux.Vars(r)
	name := values["name"]
	if name == "" {
		name = "unknown"
	}
	w.Write([]byte(fmt.Sprintf("hello %s\n", name)))
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{name}", echo).Methods("GET")
	router.HandleFunc("/", echo).Methods("GET")

	fmt.Println("listening on port 8000")
	http.ListenAndServe(":8000", router)
}
