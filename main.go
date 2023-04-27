package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

// func transfer(destination io.WriteCloser, source io.ReadCloser) {

// 	defer destination.Close()
// 	defer source.Close()
// 	io.Copy(destination, source)
// }

func transfer(dst net.Conn, src net.Conn) {

	fmt.Println("Source Local Addr: ", src.LocalAddr())
	fmt.Println("Source Remote Addr: ", src.RemoteAddr())
	fmt.Println("*********")
	fmt.Println("Destination Local Addr: ", dst.LocalAddr())
	fmt.Println("Destination Remote Addr: ", dst.RemoteAddr())
	fmt.Println("--------")

	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
func main() {
	var pemPath string
	flag.StringVar(&pemPath, "pem", "server.pem", "path to pem file")
	var keyPath string
	flag.StringVar(&keyPath, "key", "server.key", "path to key file")
	var proto string
	flag.StringVar(&proto, "proto", "https", "Proxy protocol (http or https)")
	flag.Parse()
	if proto != "http" && proto != "https" {
		log.Fatal("Protocol must be either http or https")
	}
	server := &http.Server{
		Addr: ":8090",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	if proto == "http" {
		log.Fatal(server.ListenAndServe())
	} else {
		log.Fatal(server.ListenAndServeTLS(pemPath, keyPath))
	}
}

// func main() {

// 	server := http.Server{
// 		Addr: "10.145.66.78:8090",
// 		Handler: http.HandlerFunc(
// 			func(w http.ResponseWriter, r *http.Request) {
// 				forward(w, r)
// 			},
// 		),
// 		// TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
// 	}

// 	server.ListenAndServe()
// }

// func forward(w http.ResponseWriter, r *http.Request) {
// 	var serverSideCon net.Conn
// 	var clientSideCon net.Conn
// 	var err error

// 	// get tcp connection to server
// 	fmt.Println("Dialing ", r.Host)
// 	if serverSideCon, err = net.Dial("tcp", r.Host); err != nil {
// 		log.Printf("Cannot Dial: %s \n", r.Host)
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	hijacker, ok := w.(http.Hijacker)
// 	if !ok {
// 		return
// 	}

// 	clientSideCon, _, err = hijacker.Hijack()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	// var cbuf bytes.Buffer
// 	// // clientSideCon.Read(cbuf)
// 	// io.Copy(&cbuf, clientSideCon)
// 	// fmt.Printf("From client side %s\n", cbuf.String())

// 	// var sbuf bytes.Buffer
// 	// // serverSideCon.Read(sbuf)
// 	// io.Copy(&sbuf, serverSideCon)
// 	// fmt.Printf("From server side %s\n", sbuf)

// 	// defer serverSideCon.Close()
// 	// defer clientSideCon.Close()

// 	go transfer(serverSideCon, clientSideCon)
// 	go transfer(clientSideCon, serverSideCon)
// 	// get connection to client

// }

// func transfer(destination io.WriteCloser, source io.ReadCloser) {
// 	defer destination.Close()
// 	defer source.Close()
// 	io.Copy(destination, source)
// }
