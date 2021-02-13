package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		fmt.Print("Please provide one or more backends to load balance")
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	fmt.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Errorf("%v", err)
	}
}
