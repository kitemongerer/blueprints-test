package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	i := 0

	log.Fatal(http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("request: " + time.Now().String())
		i++

		if i > 10 {
			println("long request: " + time.Now().String())

			time.Sleep(10 * time.Second)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("hi"))
	})))
}
