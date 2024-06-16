package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestServer(t *testing.T) {

	os.Setenv("ADDR", ":8080")
	os.Setenv("WEBHOOK_ROUTE", "http://localhost:9999/api/hooker")
	os.Setenv("RETRY_INTERVAL", "2")
	os.Setenv("MAX_RETRIES", "10")
	resCh := make(chan struct{})
	timeoutCh := time.After(10 * time.Second)

	url := "http://localhost:8080/api/verify"
	body := PostVerifyRequest{
		Id:            "123",
		Domain:        "www.example.com",
		ExpectedTXT:   "wgyf8z8cgvm2qmxpnbnldrcltvk4xqfn",
		ExpectedCNAME: "www.example.com.",
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	go startWebhook(resCh)

	server := New()
	go server.Start()

	time.Sleep(2 * time.Second)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	fmt.Println("Sending request")
	_, err = client.Do(req)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	select {
	case <-resCh:
		return
	case <-timeoutCh:
		t.Errorf("Expected domain to be verified, but got timeout instead")

	}
}

func startWebhook(resCh chan<- struct{}) {
	fmt.Println("Starting a hooker")
	http.HandleFunc("/api/hooker", func(w http.ResponseWriter, r *http.Request) {
		// Read POST request
		close(resCh)
	})
	err := http.ListenAndServe("localhost:9999", nil)
	if err != nil {
		panic(err)
	}
}
