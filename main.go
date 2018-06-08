package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this call was relayed by the reverse proxy")
	}))
	defer backendServer.Close()

	rpURL, err := url.Parse(backendServer.URL)
	if err != nil {
		log.Fatal(err)
	}

	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)

	director := func(req *http.Request) {
		var proxyURL *url.URL
		if randGen.Intn(100) > 50 {
			proxyURL = rpURL
		} else {
			proxyURL, _ = url.Parse("https://now.httpbin.org/")
		}
		req.Host = proxyURL.Host
		req.URL.Scheme = proxyURL.Scheme
		req.URL.Host = proxyURL.Host

		go reportAccessTime(proxyURL)
	}
	reverseProxy := &httputil.ReverseProxy{Director: director}

	frontendProxy := httptest.NewServer(reverseProxy)
	defer frontendProxy.Close()

	resp, err := http.Get(frontendProxy.URL)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", b)

}
