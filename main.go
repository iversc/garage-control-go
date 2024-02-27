package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var authCode []byte 
var hueUser string

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

func deactivate() {
    time.Sleep(1 * time.Second)
    cmd := exec.Command("gpio", "-g", "write", "4", "0")
    if err := cmd.Run(); err != nil {
        fmt.Println("gpio off call failed.")
        return
    }

    fmt.Println("Button reset.")
}

func activate() {
    cmd := exec.Command("gpio", "-g", "write", "4", "1")
    if err := cmd.Run(); err != nil {
        fmt.Println("gpio on call failed.")
        return
    }

    go deactivate()

    fmt.Println("Button triggered.")
}

func shutdown(sdType string) {
    if sdType == "-h" {
        fmt.Println("Shutting down Pi.")
    } else if sdType == "-r" {
        fmt.Println("Rebooting Pi.") 
    } else {
        fmt.Printf("Unknown shutdown type %s.\n",sdType)
        return
    }

    cmd := exec.Command("shutdown", sdType, "now")
    if err := cmd.Run(); err != nil {
        fmt.Println("shutdown call failed.")
        return
    }
}

func switchLights(enabled string) {
    body := `{ "on": {"on": ` + enabled + "}}"
    reqURL := "https://philips-hue/clip/v2/resource/grouped_light/b6faa6e6-dba5-4e11-81e1-bf92f67b9280"
    bodyReader := bytes.NewReader([]byte(body))

    req, err := http.NewRequest(http.MethodPut, reqURL, bodyReader)
    if err != nil {
        fmt.Printf("error making http request: %s\n", err)
        os.Exit(1)
    }

    req.Header.Set("hue-application-key", hueUser)
    req.Header.Set("Content-Type", "application/json")

    customTransport := http.DefaultTransport.(*http.Transport).Clone()
    customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
    client := &http.Client{Transport: customTransport}

    res, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error making request: %s\n", err)
        os.Exit(1)
    }

    resBody, err := io.ReadAll(res.Body)
    if err != nil {
        fmt.Printf("Unable to read response: %s\n", err)
        os.Exit(1)
    }

    fmt.Printf("Response: %s\n", resBody)
}

func runCommand(cmd string, w http.ResponseWriter) {
	var msg string
	if cmd == "activate" {
		fmt.Println("Door activate command received.")
        activate()
		msg = "Door Activated"
	} else if cmd == "shutdown" {
		fmt.Println("Shutdown command received.")
        shutdown("-h")
		msg = "Shutting down"
	} else if cmd == "reboot" {
		fmt.Println("Reboot command received.")
        shutdown("-r")
		msg = "Rebooting..."
	} else if cmd == "lightson" {
		fmt.Println("Lights on command received.")
        switchLights("true")
		msg = "Lights On"
	} else if cmd == "lightsoff" {
		fmt.Println("Lights off command received.")
        switchLights("false")
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
