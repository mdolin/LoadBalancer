package balancer

import (
	"fmt"
	"io"
	"loadBalancer/types"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type BalancerStrategy interface {
	GetNextServer() (*types.Server, error)
}

type LoadBalancer struct {
	strategy BalancerStrategy
}

func NewLoadBalancer(strategy BalancerStrategy) LoadBalancer {
	return LoadBalancer{
		strategy: strategy,
	}
}

func (lb *LoadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	server, err := lb.strategy.GetNextServer()
	if err != nil {
		http.Error(w, "No servers available", http.StatusServiceUnavailable)
		return
	}

	targetURL, err := url.Parse(server.URL)
	if err != nil {
		http.Error(w, "Invalid server URL", http.StatusInternalServerError)
		return
	}

	targetPath := strings.TrimRight(targetURL.String(), "/") + r.URL.Path
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetPath, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	for k, values := range r.Header {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}

	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach server", http.StatusBadGateway)
		return
	}

	byteResp, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Failed to read server response", http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	for k, values := range res.Header {
		for _, value := range values {
			w.Header().Add(k, value)
		}
	}

	w.WriteHeader(res.StatusCode)

	fmt.Fprint(w, string(byteResp))
}
