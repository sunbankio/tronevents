package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hibiken/asynqmon"
	"github.com/hibiken/asynq"
)

func main() {
	var (
		redisAddr = flag.String("redis_addr", "localhost:6379", "Redis server address")
		port      = flag.Int("port", 6382, "Port for asynqmon server to listen on")
	)
	flag.Parse()

	// Create asynqmon handler
	h := asynqmon.New(asynqmon.Options{
		RootPath: "/ui", // Root path for asynqmon UI
		RedisConnOpt: asynq.RedisClientOpt{
			Addr: *redisAddr,
		},
	})

	// Start the server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting asynqmon server on %s", addr)
	log.Printf("Redis server address: %s", *redisAddr)
	log.Printf("Access the UI at: http://localhost%s/ui", addr)

	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}