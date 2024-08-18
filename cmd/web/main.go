package main

import (
	"embed"
	"fmt"

	"html/template"
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

//go:embed index.html
var indexBuff embed.FS

//go:embed partials/redirects.html
var redirectsBuff embed.FS

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// index page

	data, err := indexBuff.ReadFile("index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	t, err := template.New("index").Parse(string(data))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	t.Execute(w, nil)
	return
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	// ping handler for healthchecks etc
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

/***
*
* Redirect Inspection Handler
*
***/

type Redirect struct {
	Url     string              `json:"url"`
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
}

func redirectPolicyFuncWrapper(redirects *[]Redirect) func(req *http.Request, via []*http.Request) error {

	return func(req *http.Request, via []*http.Request) error {
		// stop after 20 redirects (just for kicks)
		// for each redirect, check the status code and add to a list of Redirects

		if len(via) >= 20 {
			return fmt.Errorf("stopped after 20 redirects")
		}
		r := &Redirect{
			Url:     req.URL.String(),
			Status:  req.Response.StatusCode,
			Headers: req.Response.Header,
		}

		*redirects = append(*redirects, *r)

		log.Printf("Redirecting to %v", r)
		log.Printf("Redirecting to %v", req.URL)

		return nil
	}
}

// take a url query parameter and trace the redirects
// return a json response with the redirect chain
// to do this we create a custom client with a redirect policy function.
// the function we pass is retunred by a closure that has access to a
// slice of Redirects. Each time the function is called, it appends the
// current redirect to the slice.
// we then return a marshalled version of the slice as the response.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	target_url := r.URL.Query().Get("url")
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

	log.Printf("Tracing url redirects starting with %v", target_url)

	// create a slice to store the redirects
	redirects := []Redirect{}
	// instantiate a client
	client := &http.Client{
		CheckRedirect: redirectPolicyFuncWrapper(&redirects),
	}

	// make the request
	resp, err := client.Get(target_url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	data, err := redirectsBuff.ReadFile("partials/redirects.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	t, err := template.New("redirects").Parse(string(data))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("X-Request-Sid", "2koODokX4qOX88Cm78iTT42SDcl")
	w.WriteHeader(http.StatusOK)
	t.Execute(w, redirects)

	return
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
	router.Get("/redirect", redirectHandler)
	router.Get("/", indexHandler)

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
