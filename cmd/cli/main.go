// go-proxy cli
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

const usage string = `Usage:

    go-proxy [--server localhost:8080] --url http://example.com

CLI utility to proxy requests to a given url. relies on the server included in
this project. Expects the server to be deployed at a known location. the server
is configurable via the --server flag.

The server can also be configured in a yaml file at ~/.config/go-proxy.

Options:
    -s, --server  Specify which go-proxy server to use.
    -u, --url     Specify the URL to fetch.
`

func proxyRequest(server string, address string) {
	// validate server and validate url
	serverUrl, err := url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}

	addressUrl, err := url.Parse(address)
	if err != nil {
		log.Fatal(err)
	}
	if addressUrl.Scheme != "http" && addressUrl.Scheme != "https" {
		log.Fatal("The scheme of the URL should be http or https.")
	}

	// compose url for proxy request
	q := serverUrl.Query()
	q.Set("url", addressUrl.String())
	requestUrl, _ := url.Parse(serverUrl.String())
	requestUrl.RawQuery = q.Encode()
	requestUrl.Path = "/proxy"

	fmt.Println("Preparing to make request: %v", requestUrl.String())

	// make request
	resp, err := http.Get(requestUrl.String())
	if err != nil {
		log.Fatal(err)
	}
	// TODO: log redirects
	// TODO: print headers

	// print body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	fmt.Print(string(body))
}

func main() {
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }
	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	// TODO: add config file loading and env var loading (GO_PROXY_*)
	// var (
	// 	configServer string
	// )

	var (
		flagServer string
		flagProxy  bool
		flagURL    string
	)
	flag.StringVar(&flagServer, "server", "http://localhost:8080", "The go-proxy server to proxy requests via.")
	flag.BoolVar(&flagProxy, "p", false, "Proxy a request to the given url.")
	flag.BoolVar(&flagProxy, "proxy", true, "Proxy a request to the given url.")
	flag.StringVar(&flagURL, "u", "https://example.com", "The requested url.")
	flag.StringVar(&flagURL, "url", "https://example.com", "The requested url.")
	flag.Parse()

	if flagProxy {
		proxyRequest(flagServer, flagURL)
	}
}
