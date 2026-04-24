package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"secure-api-gateway/internal/logger"
)

type Proxy struct {
	target  *url.URL
	engine  *httputil.ReverseProxy
	isAlive bool
	mu      sync.RWMutex
}

func NewProxy(targetURL string) (*Proxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	p := &Proxy{
		target:  target,
		engine:  httputil.NewSingleHostReverseProxy(target),
		isAlive: true,
	}

	p.engine.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.setAlive(false)

		log := logger.FromContext(r.Context())
		log.Error("Backend is unreachable", "err", err)

		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Error: Backend is temporarily unavailable. Retrying..."))
	}

	go p.healthCheck()

	return p, nil
}

func (p *Proxy) setAlive(status bool) {
	p.mu.Lock()
	p.isAlive = status
	p.mu.Unlock()
}

func (p *Proxy) IsAlive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isAlive
}

func (p *Proxy) healthCheck() {
	for {
		time.Sleep(5 * time.Second)

		conn, err := net.DialTimeout("tcp", p.target.Host, 2*time.Second)

		if err != nil {
			if p.IsAlive() {
				logger.Log.Warn("Backend status changed: ALIVE -> DEAD", "target", p.target.Host)
				p.setAlive(false)
			}
			continue
		}

		conn.Close()

		if !p.IsAlive() {
			logger.Log.Debug("Backend status changed: DEAD -> ALIVE", "target", p.target.Host)
			p.setAlive(true)
		}
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Host = p.target.Host
	p.engine.ServeHTTP(w, r)
}
