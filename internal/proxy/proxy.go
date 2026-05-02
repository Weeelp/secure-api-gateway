package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"secure-api-gateway/internal/logger"
)

type Proxy struct {
	backends []*Backend
	current  uint32
}

type Backend struct {
	target  *url.URL
	engine  *httputil.ReverseProxy
	isAlive bool
	mu      sync.RWMutex
}

func (p *Proxy) IsAnyAlive() bool {
	for _, back := range p.backends {
		if back.IsAlive() {
			return true
		}
	}
	return false
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n := atomic.AddUint32(&p.current, 1)
	lenBack := uint32(len(p.backends))

	for i := uint32(0); i < lenBack; i++ {
		idx := (n + i) % lenBack
		peer := p.backends[idx]

		if peer.IsAlive() {
			peer.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "Service Unvailable: Np healthy backend", http.StatusServiceUnavailable)

}

func NewProxy(targetURLs []string) (*Proxy, error) {
	var backends []*Backend
	for _, targetURL := range targetURLs {
		target, err := url.Parse(targetURL)
		if err != nil {
			return nil, err
		}

		b := &Backend{
			target:  target,
			engine:  httputil.NewSingleHostReverseProxy(target),
			isAlive: true,
		}

		b.engine.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			b.setAlive(false)

			log := logger.FromContext(r.Context())
			log.Error("Backend is unreachable", "err", err)

			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Error: Backend is temporarily unavailable."))
		}

		go b.healthCheck()

		backends = append(backends, b)
	}

	return &Proxy{
		backends: backends,
		current:  0,
	}, nil
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.isAlive
}

func (b *Backend) setAlive(status bool) {
	b.mu.Lock()
	b.isAlive = status
	b.mu.Unlock()
}

func (b *Backend) healthCheck() {
	for {
		time.Sleep(5 * time.Second)

		conn, err := net.DialTimeout("tcp", b.target.Host, 2*time.Second)

		if err != nil {
			if b.IsAlive() {
				logger.Log.Warn("Backend status changed: ALIVE -> DEAD", "target", b.target.Host)
				b.setAlive(false)
			}
			continue
		}

		conn.Close()

		if !b.IsAlive() {
			logger.Log.Debug("Backend status changed: DEAD -> ALIVE", "target", b.target.Host)
			b.setAlive(true)
		}
	}
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Host = b.target.Host
	b.engine.ServeHTTP(w, r)
}
