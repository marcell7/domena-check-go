package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	Config Config
	Mux    *http.ServeMux
}

type Config struct {
	Addr          string
	WebhookRoute  string
	MaxRetries    int
	RetryInterval int
}

func New() *Server {
	maxRetries, err := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	if err != nil {
		panic(err)
	}
	retryInterval, err := strconv.Atoi(os.Getenv("RETRY_INTERVAL"))
	if err != nil {
		panic(err)
	}
	config := Config{
		Addr:          os.Getenv("ADDR"),
		WebhookRoute:  os.Getenv("WEBHOOK_ROUTE"),
		MaxRetries:    maxRetries,
		RetryInterval: retryInterval,
	}
	s := &Server{}
	mux := http.NewServeMux()
	s.Mux = mux
	s.Config = config
	mux.HandleFunc("/api/verify", s.handlerVerify)
	return s
}

func (s *Server) Start() {
	fmt.Println("Server listening on 8080")
	http.ListenAndServe(s.Config.Addr, s.Mux)
}

func (s *Server) handlerVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var body PostVerifyRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			fmt.Println(err)
			s.returnJSON(w, 401, map[string]string{
				"status": "failed",
			})
			return
		}
		go s.startDomainCheckingJob(body)
		s.returnJSON(w, 200, map[string]string{
			"status": "ok",
		})

	}
}

func (s *Server) startDomainCheckingJob(body PostVerifyRequest) error {
	txtVerified := false
	cnameVerified := false
	retries := 0
	ticker := time.NewTicker(time.Duration(s.Config.RetryInterval) * time.Second)
	done := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			func() error {
				if !txtVerified {
					txtVerified = s.verifyTXTRecord(body.Domain, body.ExpectedTXT)
				}
				if !cnameVerified {
					cnameVerified = s.verifyCNAMERecord(body.Domain, body.ExpectedCNAME)
				}
				if txtVerified && cnameVerified {
					if err := s.sendResultToWebhook(body); err != nil {
						return err
					}
					close(done)
					return nil
				}
				if retries > s.Config.MaxRetries {
					close(done)
					return fmt.Errorf("max retry limit exceeded")
				}
				retries += 1
				return nil
			}()

		case <-done:
			ticker.Stop()
			return nil
		}
	}
}

func (s *Server) verifyTXTRecord(domain string, expectedRecord string) bool {
	records, err := net.LookupTXT(domain)
	if err != nil {
		return false
	}
	for _, record := range records {
		if record == expectedRecord {
			return true
		}
	}
	return false
}

func (s *Server) verifyCNAMERecord(domain string, expectedRecord string) bool {
	record, err := net.LookupCNAME(domain)
	if err != nil {
		return false
	}
	if record == expectedRecord {
		return true
	} else {
		return false
	}
}

func (s *Server) sendResultToWebhook(oldBody PostVerifyRequest) error {
	body := &PostWebhookRequest{
		Id:            oldBody.Id,
		Domain:        oldBody.Domain,
		AcquiredTXT:   oldBody.ExpectedTXT,
		AcquiredCNAME: oldBody.ExpectedCNAME,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", s.Config.WebhookRoute, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) returnJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}
