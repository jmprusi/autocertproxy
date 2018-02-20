package main

import (
	"autocertproxy/redisCache"
	"crypto/tls"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type AutocertConfig struct {
	Debug       bool
	HTTPPort    int
	HTTPSPort   int
	ProxyURL    string
	RedisURL    string
	RenewBefore int
}

func main() {

	parsedUrl, err := url.Parse("https://a:a@localhost:9091")

	if err != nil {
		panic("ProxyURL config is invalid.")
	}

	cache, err := rediscache.New("redis://localhost:6379")

	if err != nil {
		panic("Couldn't initialize redis cache.")
	}

	m := &autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		Cache:       cache,
		HostPolicy:  autocert.HostWhitelist("dfb0de6d.ngrok.io"),
		RenewBefore: 30,
		Client:      nil,
		Email:       "",
		ForceRSA:    false,
	}
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: parsedUrl.Scheme,
		User:   parsedUrl.User,
		Host:   parsedUrl.Host,
		Path:   parsedUrl.Path,
	})

	go http.ListenAndServe(":http", m.HTTPHandler(nil))

	s := &http.Server{
		Addr:      ":https",
		Handler:   proxy,
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}
	log.Fatal(s.ListenAndServeTLS("", ""))

}
