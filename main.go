package main

import (
	"fmt"
	"log"
	"net/http"
	"ticket-reservation/handler"
	"ticket-reservation/service"
	"ticket-reservation/store"
)

func main() {
	s := store.New()
	svc := service.NewReservationService(s)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":8080"
	fmt.Printf("ticket reservation server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
