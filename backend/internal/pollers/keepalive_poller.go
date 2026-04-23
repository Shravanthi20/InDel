package pollers

import (
	"log"
	"net/http"
	"time"
)

// KeepAlivePoller periodically pings backend services to keep them alive on Render.
type KeepAlivePoller struct {
	ServiceURLs []string
}

func (p *KeepAlivePoller) Start() {
	go func() {
		p.pingAll()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			p.pingAll()
		}
	}()
}

func (p *KeepAlivePoller) pingAll() {
	for _, url := range p.ServiceURLs {
		go func(serviceURL string) {
			resp, err := http.Get(serviceURL)
			if err != nil {
				log.Printf("[KeepAlivePoller] Failed to ping %s: %v", serviceURL, err)
				return
			}
			resp.Body.Close()
			log.Printf("[KeepAlivePoller] Pinged %s, status: %s", serviceURL, resp.Status)
		}(url)
	}
}
