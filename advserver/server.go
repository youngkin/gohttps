// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	help := flag.Bool("help", false, "Optional, prints usage info")
	host := flag.String("host", "", "Required flag, must be the hostname that is resolvable via DNS, or 'localhost'")
	port := flag.String("port", "443", "The https port, defaults to 443")
	serverCert := flag.String("srvcert", "", "Required, the name of the server's certificate file")
	caCert := flag.String("cacert", "", "Required, the name of the CA that signed the client's certificate")
	srcKey := flag.String("srvkey", "", "Required, the file name of the server's private key file")
	certOpt := flag.Int("certopt", 0, "Optional, specifies the option for authenticating a client via certificate")
	flag.Parse()

	usage := `usage:
	
simpleserver -host <hostname> -srvcert <serverCertFile> -cacert <caCertFile> -srvkey <serverPrivateKeyFile> [-port <port> -certopt <certopt> -help]
	
Options:
  -help       Prints this message
  -host       Required, a DNS resolvable host name
  -srvcert    Required, the name the server's certificate file
  -cacert     Required, the name of the CA that signed the client's certificate
  -srvkey     Required, the name the server's key certificate file
  -port       Optional, the https port for the server to listen on
  -certopt    Optional, specifies the option for authenticating a client via certificate:
			  0 - certificate not required, 
			  1 - request a certificate but it's not required,
			  2 - require any client certificate
			  3 - if provided, verify the client certificate is authorized
			  4 - require certificate and verify it's authorized`

	if *help == true {
		fmt.Println(usage)
		return
	}
	if *host == "" || *serverCert == "" || *caCert == "" || *srcKey == "" {
		log.Fatalf("One or more required fields missing:\n%s", usage)
	}

	if *certOpt < 0 || *certOpt > 4 {
		log.Fatalf("Invalid value %d, provided for 'certopt' flag. It must be a number between 0 and 4 inclusive.\n%s", *certOpt, usage)
	}

	server := &http.Server{
		Addr:         ":" + *port,
		ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		WriteTimeout: 10 * time.Second,
		TLSConfig:    getTLSConfig(*host, *caCert, tls.ClientAuthType(*certOpt)),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for host %s from IP address %s and X-FORWARDED-FOR %s",
			r.Method, r.Host, r.RemoteAddr, r.Header.Get("X-FORWARDED-FOR"))
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			body = []byte(fmt.Sprintf("error reading request body: %s", err))
		}
		resp := fmt.Sprintf("Hello, %s from Advanced Server!", body)
		w.Write([]byte(resp))
		log.Printf("Advanced Server: Sent response %s", resp)
	})

	log.Printf("Starting HTTPS server on host %s and port %s", *host, *port)
	if err := server.ListenAndServeTLS(*serverCert, *srcKey); err != nil {
		log.Fatal(err)
	}
}

func getTLSConfig(host, caCertFile string, certOpt tls.ClientAuthType) *tls.Config {
	var caCert []byte
	var err error
	var caCertPool *x509.CertPool
	if certOpt > tls.RequestClientCert {
		caCert, err = ioutil.ReadFile(caCertFile)
		if err != nil {
			log.Fatal("Error opening cert file", caCertFile, ", error ", err)
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	return &tls.Config{
		ServerName: host,
		// ClientAuth: tls.NoClientCert,				// Client certificate will not be requested and it is not required
		// ClientAuth: tls.RequestClientCert,			// Client certificate will be requested, but it is not required
		// ClientAuth: tls.RequireAnyClientCert,		// Client certificate is required, but any client certificate is acceptable
		// ClientAuth: tls.VerifyClientCertIfGiven,		// Client certificate will be requested and if present must be in the server's Certificate Pool
		// ClientAuth: tls.RequireAndVerifyClientCert,	// Client certificate will be required and must be present in the server's Certificate Pool
		ClientAuth: certOpt,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12, // TLS versions below 1.2 are considered insecure - see https://www.rfc-editor.org/rfc/rfc7525.txt for details
	}
}
