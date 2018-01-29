package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"log"
	"golang.org/x/crypto/acme/autocert"
	"crypto/tls"
	)

func main() {

	log.Info("Starting AutoCertProxy")

	m := &autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		Cache:       autocert.DirCache("secret-dir"),
		HostPolicy:  autocert.HostWhitelist("keepalive.io"),
		RenewBefore: 0,
		Email:       "",
		ForceRSA:    false,
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme:     "http",
		Host:       "localhost:9091",
	})

	go http.ListenAndServe(":http", m.HTTPHandler(nil))

	s := &http.Server{
		Addr:              ":https",
		Handler:           proxy,
		TLSConfig:         &tls.Config{GetCertificate: m.GetCertificate},
	}

	log.Fatal(s.ListenAndServeTLS("", ""))

}