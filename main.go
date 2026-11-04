package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"math/big"
	"net/http"
	"net/url"
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

func validURL(raw string) bool {
	if len(raw) > 2048 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

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

func store(ctx context.Context, target string) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		c, err := newCode()
		if err != nil {
			return "", err
		}
		if !reserved[c] {
			res, err := db.ExecContext(
				ctx,
				"INSERT INTO links (code, url) VALUES ($1, $2) ON CONFLICT (code) DO NOTHING",
				c, target)
			if err != nil {
				return "", err
			}

			n, err := res.RowsAffected()
			if err != nil {
				return "", err
			}
			if n == 1 {
				return c, nil
			}
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
	if !validURL(req.URL) {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	code, err := store(r.Context(), req.URL)
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

	var target string
	err := db.QueryRowContext(r.Context(),
		"UPDATE links SET clicks = clicks + 1 WHERE code = $1 RETURNING url",
		code).Scan(&target)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, target, http.StatusFound)
}

var reserved = map[string]bool{"health": true}
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("pgx", "postgres://nhh4075@localhost:5432/url_shortener")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /health", health)
	http.HandleFunc("POST /shorten", shorten)
	http.HandleFunc("GET /{code}", redirect)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
