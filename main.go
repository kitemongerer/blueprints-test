package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("SLOW_HEALTHCHECK") != "" {
		println("starting with slow healthcheck")
		slowHealthcheck(port)
	} else {
		println("starting with default server")
		defaultServer(port)
	}
}

func defaultServer(port string) {
	log.Fatal(http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("received request at %s\n", r.URL.Path)

		if strings.Contains(r.URL.Path, "server-error") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if strings.Contains(r.URL.Path, "panic") {
			panic(string(debug.Stack()))
			return
		}
		if strings.Contains(r.URL.Path, "exit") {
			os.Exit(17)
		}
		if strings.Contains(r.URL.Path, "oom") {
			go oom()
			w.Write([]byte("started oom loop"))
		}
		w.Write([]byte("hi"))
	})))
}

func slowHealthcheck(port string) {
	i := 0

	log.Fatal(http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%d request (%s) at %s\n", i, r.URL.Path, time.Now().String())
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

func oom() {
	buf := bytes.NewBuffer([]byte{})
	cap := 1024
	for {
		fmt.Printf("buffer capacity: %d\n", buf.Cap())
		cap += 1024
		buf.Grow(cap)
	}
}
