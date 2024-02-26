package main

import (
	"fmt"
	"io"
	"errors"
	"net/http"
	"os"
	"github.com/gorilla/mux"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got / request")
	io.WriteString(w, "Success\n")
}

func getCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("got /commmand/%s request\n", vars["cmd"])
	fmt.Fprintf(w, "Command %s\n", vars["cmd"])
}

func main() {
	fmt.Println("Setting up server...")
	r := mux.NewRouter()
//	http.HandleFunc("/", getRoot)
	r.HandleFunc("/command/{cmd}", getCommand)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.Handle("/command/", r)

	err := http.ListenAndServe("localhost:3000", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server shutting down")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
