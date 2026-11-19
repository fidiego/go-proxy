package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"go-proxy/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//
// Request Handlers
//

//go:embed index.html partials/*.html
var assets embed.FS

// templates are parsed once at startup; names are the file base names
// ("index.html", "redirects.html").
var templates = template.Must(template.ParseFS(assets, "index.html", "partials/*.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// index page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		log.Printf("index template error: %v", err)
	}
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

const maxRedirects = 20

// Hop is a single response in a redirect chain: the URL that was requested,
// the status it answered with, and (for redirects) where it pointed next.
type Hop struct {
	Url      string      `json:"url"`
	Status   int         `json:"status"`
	Location string      `json:"location,omitempty"`
	Headers  http.Header `json:"headers"`
	Final    bool        `json:"final"`
}

// StatusClass returns a short label used by the UI to colour the status pill.
func (h Hop) StatusClass() string {
	switch {
	case h.Status >= 500:
		return "err"
	case h.Status >= 400:
		return "warn"
	case h.Status >= 300:
		return "redir"
	case h.Status >= 200:
		return "ok"
	default:
		return ""
	}
}

// RedirectTrace is what the redirects partial renders.
type RedirectTrace struct {
	Target    string
	Hops      []Hop // every response in order; the last one is the final response
	Redirects int   // number of redirects followed (len(Hops) - 1 when successful)
	Elapsed   time.Duration
	Error     string
}

func (t RedirectTrace) ElapsedMs() int64 { return t.Elapsed.Milliseconds() }

func (t RedirectTrace) FinalHop() *Hop {
	if len(t.Hops) == 0 {
		return nil
	}
	return &t.Hops[len(t.Hops)-1]
}

func redirectPolicyFuncWrapper(hops *[]Hop) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		// called before following each redirect. req is the *next* request,
		// req.Response is the redirect response, via[len-1] is the request that
		// received it.
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		prev := via[len(via)-1]
		hop := Hop{
			Url:      prev.URL.String(),
			Status:   req.Response.StatusCode,
			Location: req.URL.String(),
			Headers:  req.Response.Header,
		}
		*hops = append(*hops, hop)
		log.Printf("redirect %d: %s -> %d -> %s", len(via), hop.Url, hop.Status, hop.Location)
		return nil
	}
}

func renderTrace(w http.ResponseWriter, status int, trace RedirectTrace) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := templates.ExecuteTemplate(w, "redirects.html", trace); err != nil {
		log.Printf("redirects template error: %v", err)
	}
}

// take a url query parameter and trace the redirects, returning an html
// partial with the full chain (every hop plus the final response). the final
// response is always included, so a url with zero redirects still shows its
// status and headers.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	target_url := r.URL.Query().Get("url")
	trace := RedirectTrace{Target: target_url}

	// validate url (http, https)
	if len(target_url) == 0 {
		trace.Error = "No URL provided. Enter a URL to trace."
		renderTrace(w, http.StatusBadRequest, trace)
		return
	}
	parsed, err := url.Parse(target_url)
	if err != nil {
		trace.Error = "Could not parse that URL."
		renderTrace(w, http.StatusBadRequest, trace)
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		trace.Error = "URL must begin with http:// or https://."
		renderTrace(w, http.StatusBadRequest, trace)
		return
	}

	log.Printf("Tracing url redirects starting with %v", target_url)

	hops := []Hop{}
	client := &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: redirectPolicyFuncWrapper(&hops),
	}

	start := time.Now()
	resp, err := client.Get(target_url)
	trace.Elapsed = time.Since(start)

	// when CheckRedirect returns an error the client still hands back the last
	// response (with its body closed), so we can show the partial chain.
	if resp != nil {
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		hops = append(hops, Hop{
			Url:     resp.Request.URL.String(),
			Status:  resp.StatusCode,
			Headers: resp.Header,
			Final:   err == nil,
		})
	}
	// every hop is a redirect except a successful final response
	trace.Hops = hops
	trace.Redirects = len(hops)
	if last := trace.FinalHop(); last != nil && last.Final {
		trace.Redirects--
	}

	if err != nil {
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		trace.Error = err.Error()
		renderTrace(w, http.StatusBadGateway, trace)
		return
	}

	renderTrace(w, http.StatusOK, trace)
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
