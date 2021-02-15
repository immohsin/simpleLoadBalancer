package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

func GetAttemptsFromContext(r *http.Request) int {
	ctx := r.Context()
	if retry, ok := ctx.Value(Attempts).(int); ok {
		return retry
	}
	return 0
}

func GetRetryFromContext(r *http.Request) int {
	ctx := r.Context()
	if retry, ok := ctx.Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func healthCheck() {
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			fmt.Print("Starting health check...")
			serverPool.HealthCheck()
			fmt.Print("Health check complete...")
		}
	}
}
func isBackendAlive(u *url.URL) bool {
	timeOut := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeOut)
	if err != nil {
		fmt.Print("Backend unreachable, error: ", err)
		return false
	}
	_ = conn.Close()
	return true

}
