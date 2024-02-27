package main

import (
	"fmt"
	"io"
	"errors"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"time"
	"strings"
)

var authCode []byte 

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got / request")
	io.WriteString(w, "Success\n")
}

func checkAuth(clientAuthCode string) bool {
	cacBytes, err := hex.DecodeString(clientAuthCode)
	if err != nil {
		fmt.Println("Unable to decode CAC.")
		return false
	}

	currentTimeWindow := time.Now().Unix() / 30
	ctwBytes := []byte(fmt.Sprint(currentTimeWindow))

	mac := hmac.New(sha1.New, authCode)
	mac.Write(ctwBytes)
	expected := mac.Sum(nil)

	if hmac.Equal(expected, cacBytes) {
		return true
	}

	fmt.Println("HMAC failed, checking previous time window")
	currentTimeWindow -= 1
	ctwBytes = []byte(fmt.Sprint(currentTimeWindow))

	mac.Reset()
	mac.Write(ctwBytes)
	expected = mac.Sum(nil)

	if hmac.Equal(expected, cacBytes) {
		return true
	}

	return false
}

func runCommand(cmd string, w http.ResponseWriter) {
	var msg string
	if cmd == "activate" {
		fmt.Println("Door activate command received.")
		msg = "Door Activated"
	} else if cmd == "shutdown" {
		fmt.Println("Shutdown command received.")
		msg = "Shutting down"
	} else if cmd == "reboot" {
		fmt.Println("Reboot command received.")
		msg = "Rebooting..."
	} else if cmd == "lightson" {
		fmt.Println("Lights on command received.")
		msg = "Lights On"
	} else if cmd == "lightsoff" {
		fmt.Println("Lights off command received.")
		msg = "Lights Off"
	} else {
		fmt.Printf("Unknown command '%s' received.\n", cmd)
		msg = "Unknown Command"
		w.WriteHeader(http.StatusBadRequest)
	}

	io.WriteString(w, msg)	
}

func getCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("got /commmand/%s request\n", vars["cmd"])
	clientAuthCode := strings.Split(r.Header["Authorization"][0], " ")[1]
	if !checkAuth(clientAuthCode) {
		fmt.Println("Invalid auth code from client.")
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Invalid authorization code.\n")
		return
	}

	runCommand(vars["cmd"], w)
}

func main() {
	myAuthCode, err := os.ReadFile("keyfile")
	authCode = []byte(strings.TrimSpace(string(myAuthCode)))
	if err != nil {
		fmt.Println("Error reading keyfile")
		os.Exit(1)
	}
	fmt.Println("Setting up server...")
	r := mux.NewRouter()
	r.HandleFunc("/command/{cmd}", getCommand)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.Handle("/command/", r)

	err = http.ListenAndServe("localhost:3000", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server shutting down")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
