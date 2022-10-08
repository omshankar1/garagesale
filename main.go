package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// ************************************************************************
	// App Starting
	log.Println("Main: started")
	defer log.Println("Main: Completed")

	// ************************************************************************
	// Starting API Service

	api := http.Server{
		Addr:         "localhost:8000",
		Handler:      http.HandlerFunc(ListProducts),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Make a channel is listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error
	serverErrors := make(chan error, 1)

	// Start the service listening for requests
	go func() {
		log.Println("main: API listening on localhost:8000")
		serverErrors <- api.ListenAndServe()
	}()

	// Make a channel to listen for an interrupt or terminate signal from the OS
	// Use a buffered channel because signal package expects it
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// ************************************************************************
	// Shutdown

	// Blocking main and waiting for shutdown

	select {
	case err := <-serverErrors:
		log.Fatalf("error listening and Serving %s", err)
	case <-shutdown:
		log.Fatalf("main Start shutdown")
		// Give outstanding requests deadline for completion
		const timeout = time.Second * 5
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Asking listener to shutdown and load shed
		// Shutdown gracefully shuts down the server without
		// interrupting any active connection
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main: Graceful shutdown did not complete %s", err)
			// Close immediately closes all active net\.Listeners and any
			// connections in state StateNew, StateActive, or StateIdle
			err = api.Close()
		}
		if err != nil {
			log.Printf("main: Could not shutdown server gracefully %s", err)
		}
	}

}

// Product : Item that is sold
type Product struct {
	Name     string
	cost     int
	Quantity int
}

// ListProducts Gives all products a lists
func ListProducts(w http.ResponseWriter, r *http.Request) {
	list := []Product{
		{Name: "Comic Books", Cost: 75, Quantity: 50},
		{Name: "McDonand Toys", Cost: 25, Quantity: 120},
	}
}
