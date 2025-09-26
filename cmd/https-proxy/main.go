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
//	  -nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname"
package main

import (
	"cmp"
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

	log.Printf("Listening on %s", port)
	if err := http.ListenAndServeTLS(port, filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), handler); err != nil {
		log.Fatal(err)
	}
}
