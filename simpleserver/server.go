// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
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
	srvKey := flag.String("srvkey", "", "Required, the file name of the server's private key file")
	flag.Parse()

	usage := `usage:
	
simpleserver -host <hostname> -srvcert <serverCertFile> -cacert <caCertFile> -srvkey <serverPrivateKeyFile> [-port <port> -certopt <certopt> -help]
	
Options:
  -help       Prints this message
  -host       Required, a DNS resolvable host name or 'localhost'
  -srvcert    Required, the name the server's certificate file
  -srvkey     Required, the name the server's key certificate file
  -port       Optional, the https port for the server to listen on, defaults to 443
  `

	if *help == true {
		fmt.Println(usage)
		return
	}
	if *host == "" || *serverCert == "" || *srvKey == "" {
		log.Fatalf("One or more required fields missing:\n%s", usage)
	}

	server := &http.Server{
		Addr:         ":" + *port,
		ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		WriteTimeout: 10 * time.Second,
		TLSConfig:    &tls.Config{ServerName: *host},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for host %s from IP address %s and X-FORWARDED-FOR %s",
			r.Method, r.Host, r.RemoteAddr, r.Header.Get("X-FORWARDED-FOR"))
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			body = []byte(fmt.Sprintf("error reading request body: %s", err))
		}
		resp := fmt.Sprintf("Hello, %s from Simple Server!", body)
		w.Write([]byte(resp))
		log.Printf("SimpleServer: Sent response %s", resp)
	})

	log.Printf("Starting HTTPS server on host %s and port %s", *host, *port)
	if err := server.ListenAndServeTLS(*serverCert, *srvKey); err != nil {
		log.Fatal(err)
	}
}
