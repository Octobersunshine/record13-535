package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"ticket-reservation/handler"
	"ticket-reservation/lock"
	"ticket-reservation/service"
	"ticket-reservation/store"
)

func main() {
	locker, err := createLocker()
	if err != nil {
		log.Fatalf("failed to create locker: %v", err)
	}

	s := store.New(locker)
	svc := service.NewReservationService(s)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":8080"
	fmt.Printf("ticket reservation server starting on %s (locker: %s)\n", addr, os.Getenv("LOCK_TYPE"))
	log.Fatal(http.ListenAndServe(addr, mux))
}

func createLocker() (lock.Locker, error) {
	lockType := os.Getenv("LOCK_TYPE")
	if lockType == "" {
		lockType = "memory"
	}

	switch lockType {
	case "redis":
		cfg := lock.RedisConfig{
			Address:  getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		}
		return lock.NewRedisLocker(cfg)
	case "memory":
		return lock.NewMemoryLocker(), nil
	default:
		return nil, fmt.Errorf("unsupported lock type: %s", lockType)
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
