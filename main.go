package main

import (
	"fmt"
	balancer "loadBalancer/core"
	"loadBalancer/pool"
	"loadBalancer/types"
	"log"
	"net/http"
	"time"
)

var servers = []types.Server{
	{
		ID:              "1",
		Name:            "Server 1",
		Protocol:        "http",
		Host:            "localhost",
		Port:            3001,
		URL:             "http://localhost:3001",
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
	},
	{
		ID:              "2",
		Name:            "Server 2",
		Protocol:        "http",
		Host:            "localhost",
		Port:            3002,
		URL:             "http://localhost:3002",
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
	}, {
		ID:              "3",
		Name:            "Server 3",
		Protocol:        "http",
		Host:            "localhost",
		Port:            3003,
		URL:             "http://localhost:3003",
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
	},
}

func main() {
	serverPool := pool.NewServerPool()

	for _, server := range servers {
		serverPool.AddServer(&server)
	}

	rb := balancer.NewRoundRobinBalancer(serverPool)
	lb := balancer.NewLoadBalancer(rb)
	distributeLoad(3000, lb)

	select {}
}

func distributeLoad(port int, lb balancer.LoadBalancer) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", lb.Serve)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("Starting LOAD BALANCER on port %d", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Load Balancer on port %d failed: %v", port, err)
	}
}
