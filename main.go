package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	Attempts int = iota
	Retry
)

var serverPool ServerPool

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		fmt.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Request could not be processed", http.StatusServiceUnavailable)
		return
	}
	peer := serverPool.NextPeer()
	if peer == nil {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	peer.ReverseProxy.ServeHTTP(w, r)
}

func main() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		fmt.Print("Please provide one or more backends to load balance")
	}

	tokens := strings.Split(serverList, ",")

	for _, tok := range tokens {
		u, err := url.Parse(tok)
		if err != nil {
			fmt.Println("Error parsing url")
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			fmt.Printf("[%s] %s\n", u, e.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			serverPool.MarkBackendStatus(u, false)

			attempts := GetAttemptsFromContext(request)
			fmt.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))

		}
		serverPool.AddBackend(&BackEnd{
			URL:          u,
			ReverseProxy: proxy,
			Alive:        true,
		})
		fmt.Printf("Configured server: %s\n", u)

	}
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	go healthCheck()

	// fmt.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Errorf("%v", err)
	}

}
