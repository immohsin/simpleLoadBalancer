package main

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type ServerPool struct {
	backEnds []*BackEnd
	current  uint64
}

type BackEnd struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *BackEnd) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *BackEnd) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backEnds)))
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backEnds {
		alive := isBackendAlive(b.URL)
		if !alive {
			fmt.Println("Down", b.URL.Host)
		}
		b.SetAlive(alive)
	}
}

func (s *ServerPool) NextPeer() *BackEnd {
	next := s.NextIndex()
	l := len(s.backEnds) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backEnds)
		if s.backEnds[idx].IsAlive() {
			if i != next {
				atomic.AddUint64(&s.current, uint64(idx))
			}
			return s.backEnds[idx]
		}
	}
	return nil
}

func (s *ServerPool) MarkBackendStatus(serveUrl *url.URL, alive bool) {
	for _, b := range s.backEnds {
		if b.URL.String() == serveUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

func (s *ServerPool) AddBackend(backend *BackEnd) {
	s.backEnds = append(s.backEnds, backend)
}
