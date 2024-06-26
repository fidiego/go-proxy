package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"go-proxy/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//
// Request Handlers
//

// pingHandler function  
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// take a url query parameter and make a request to that url
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	target_url := r.URL.Query().Get("url")

	// check if url exists
	if len(target_url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No URL provided. Please provide a url in in the query string."))
		return
	}

	// validate url (http, https)
	url, err := url.Parse(target_url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid URL provided. Please provide a valid url in the query string.\n\nFailed to parse provided url."))
		return
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid URL provided. Please provide a valid url in the query string.\n\nMust begin with http or https."))
		return
	}

	log.Printf("Proxying request to %v", target_url)

	// make the request
	resp, err := http.Get(target_url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	w.Write([]byte(body))
}

func main() {
	log.Print(fmt.Sprintf("Preparing to start server."))

	configs := config.Configs{}
	configs.Load()

	port := configs.Port

	// create a new router
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// register routes
	router.Get("/ping", pingHandler)
	router.Get("/proxy", proxyHandler)

	// set up the server
	address := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// start listening
	log.Printf("[web] Listening on %v", address)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
