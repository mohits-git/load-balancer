// utils contains the utility helper functions for the loadbalancer
package l7lb

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

// RequestLogger logs incoming clients requests with client addr info
func HTTPRequestLogger(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, port := GetHTTPClientRemoteAddrInfo(r)

		// client info
		log.Printf("Received a request from %s:%s\n\n", ip, port)

		// http request line
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, r.Proto)
		// info/headers
		fmt.Println("Host:", r.Host)
		fmt.Println("User-Agent:", r.UserAgent())
		fmt.Println("Accept:", r.Header.Get("Accept"))
		fmt.Println()

		// continue
		next(w, r)
	})
}

// returns the ip and port from the client http request
func GetHTTPClientRemoteAddrInfo(r *http.Request) (string, string) {
	addr := r.RemoteAddr

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		addr = xff
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Println("Error while parsing address", err)
		return "", ""
	}

	return host, port
}
