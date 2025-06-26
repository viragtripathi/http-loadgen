package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a write delay and 201 Created
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, `{"status":"write ok"}`)
	})

	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a read delay and possible 429/500
		log.Printf("💥 /read called with method: %s\n", r.Method)

		delay := rand.Intn(50)
		time.Sleep(time.Duration(delay) * time.Millisecond)

		if rand.Float32() < 0.05 {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, `{"error":"rate limited"}`)
			return
		} else if rand.Float32() < 0.02 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, `{"error":"internal error"}`)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"read ok"}`)
	})

    port := os.Getenv("FAKEAPI_PORT")
    if port == "" {
        port = "8080"
    }

   	log.Printf("📡 Fake API server listening on :%s\n", port)

    http.HandleFunc("/health/alive", func(w http.ResponseWriter, r *http.Request) {
    	w.WriteHeader(http.StatusOK)
    	w.Write([]byte("ok"))
    })

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
