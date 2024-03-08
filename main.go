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
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("SLOW_HEALTHCHECK") != "" {
		println("starting with slow healthcheck")
		slowHealthcheck(port, os.Getenv("SLOW_HEALTHCHECK"))
	} else if os.Getenv("PORT_DETECTOR_TEST") != "" {
		println("starting default server, secondary server, and udp server")
		portDetectorTest()
	} else if os.Getenv("PORT_DETECTOR_TEST_2") != "" {
		println("starting port detector test 2")
		portDetectorTest2(port)
	} else {
		println("starting with default server")
		defaultServer(port)
	}
}

func defaultServer(port string) {
	log.Printf("starting http server at %s\n", port)

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

func slowHealthcheck(port string, duration string) {
	dur, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	err = http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%d seconds request (%s) at %s\n", time.Since(start).Seconds(), r.URL.Path, time.Now().String())

		if time.Since(start) > dur {
			println("long request: " + time.Now().String())

			time.Sleep(10 * time.Second)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("hi"))
	}))

	log.Fatalf("error listening on port %s: %s", port, err)
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

func portDetectorTest() {
	go defaultServer("8082")
	time.Sleep(5 * time.Second)

	udp := startUDP()
	defer udp.Close()

	tcp := startTCP()
	defer tcp.Close()

	// ensure port detector finds delayed ports
	time.Sleep(time.Minute)
	defaultServer("0")
}

func portDetectorTest2(port string) {
	go defaultServer("8082")
	defaultServer(port)
}

func startUDP() *net.UDPConn {
	s, err := net.ResolveUDPAddr("udp6", ":0")
	if err != nil {
		log.Fatalf("error resolving addr on udp: %s", err)
	}

	conn, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Fatalf("error listening on udp: %s", err)
	}
	return conn
}

func startTCP() *net.TCPListener {
	s, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		log.Fatalf("error resolving addr on udp: %s", err)
	}

	conn, err := net.ListenTCP("tcp", s)
	if err != nil {
		log.Fatalf("error listening on udp: %s", err)
	}
	return conn
}
