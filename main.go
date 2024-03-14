package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conn := startTCP(port)
	defer conn.Close()

	srv := &http.Server{Addr: conn.Addr().String(), Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})}

	log.Fatal(srv.Serve(conn))
}

func oom() {
	buf := bytes.NewBuffer([]byte{})
	cap := 1024
	for {
		fmt.Printf("buffer capacity: %d\n", buf.Cap())
		cap *= 2
		buf.Grow(cap)
	}
}

func startTCP(port string) *net.TCPListener {
	s, err := net.ResolveTCPAddr("tcp6", ":"+port)
	if err != nil {
		log.Fatalf("error resolving addr on udp: %s", err)
	}

	conn, err := net.ListenTCP("tcp6", s)
	if err != nil {
		log.Fatalf("error listening on udp: %s", err)
	}
	return conn
}
