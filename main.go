package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"sync"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code string `json:"code"`
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

var links = map[string]string{}
var reserved = map[string]bool{"health": true}
var mu sync.Mutex

func newCode() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[int(n.Int64())]
	}

	return string(b), nil
}

func store(url string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	for attempt := 0; attempt < 5; attempt++ {
		c, err := newCode()
		if err != nil {
			return "", err
		}
		if _, exists := links[c]; !exists && !reserved[c] {
			links[c] = url
			return c, nil
		}
	}
	return "", errors.New("reached max attempt")
}

func shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "empty url", http.StatusBadRequest)
		return
	}

	code, err := store(req.URL)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(shortenResponse{Code: code})
	if err != nil {
		log.Printf("encode response: %v", err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	mu.Lock()
	url, ok := links[code]
	mu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func main() {
	http.HandleFunc("GET /health", health)
	http.HandleFunc("POST /shorten", shorten)
	http.HandleFunc("GET /{code}", redirect)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
