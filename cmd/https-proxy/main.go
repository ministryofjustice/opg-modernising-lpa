// Command https-proxy provides a https proxy.
//
// Your browser will complain about the certificate if you try to visit this,
// but it really exists to allow the docker composed app to set cookies, so
// Cypress can be run from a container too.
//
// Use the following environment variables to change the behaviour:
//
//	PORT (default 8443) - the port this server will listen on
//	DIR (default .) - the directory containing the certificate and private key
//	TARGET - the URL to proxy requests to
//
// To generate the required files run something like:
//
//	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 3650 \
//		-nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname"
package main

import (
	"cmp"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	var (
		port         = ":" + cmp.Or(os.Getenv("PORT"), "8443")
		dir          = cmp.Or(os.Getenv("DIR"), ".")
		target       = os.Getenv("TARGET")
		targetURL, _ = url.Parse(target)
		handler      = httputil.NewSingleHostReverseProxy(targetURL)
	)

	if len(os.Args) == 3 && os.Args[1] == "--healthy" {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client := &http.Client{Transport: customTransport}

		resp, err := client.Get(os.Args[2])
		if err != nil {
			log.Println(err)
			os.Exit(1)
		} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			log.Println(resp.StatusCode)
			os.Exit(1)
		}

		os.Exit(0)
	}

	log.Printf("Listening on %s", port)
	if err := http.ListenAndServeTLS(port, filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), handler); err != nil {
		log.Fatal(err)
	}
}
