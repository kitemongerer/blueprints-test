package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

func main() {
	port := "8080"

	if os.Getenv("CURL_URL") != "" {
		go pollURL(os.Getenv("CURL_URL"))
	}

	println("starting with default server")
	go defaultServer(port)

	for {
		var server *http.Server
		server = &http.Server{Addr: ":0", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			if strings.Contains(r.URL.Path, "stop") {
				w.WriteHeader(http.StatusOK)
				server.Close()
			}
			if strings.Contains(r.URL.Path, "oom") {
				go oom()
				w.Write([]byte("started oom loop"))
			}
			w.Write([]byte("hi from: " + server.Addr))
		})}

		go server.ListenAndServe()
		time.Sleep(5 * time.Second)
		server.Close()
	}
}

func serveInterfaces() {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	startPort := 8000
	for idx, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			println(err)
			continue
		}
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			go serveAtAddr(fmt.Sprintf("%s:%d", ip, startPort+idx))
		}
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("received exit signal")
}

func pollURL(url string) {
	for {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("unable to poll url: %s\n", err)
			time.Sleep(time.Second)
			continue
		}

		fmt.Printf("got status: %d\n", resp.StatusCode)

		bs, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Printf("got body: %s\n", string(bs))
		}

		time.Sleep(time.Second)
	}
}

func defaultServer(port string) {
	serveAtAddr(":" + port)
}

func serveAtAddr(addr string) {
	log.Printf("starting http server at %s\n", addr)

	var server *http.Server
	server = &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if strings.Contains(r.URL.Path, "stop") {
			w.WriteHeader(http.StatusOK)
			server.Close()
		}
		if strings.Contains(r.URL.Path, "oom") {
			go oom()
			w.Write([]byte("started oom loop"))
		}
		w.Write([]byte("hi from: " + server.Addr))
	})}

	log.Println(server.ListenAndServe())
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
	go defaultServer("10001")
	defaultServer(port)
}

func portDetectorTest3(port string) {
	go defaultServer("10001")

	err := http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}))

	log.Fatalf("error listening on port %s: %s", port, err)
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

func startPorts(portsList string) {
	ports := strings.Split(portsList, ",")

	if len(ports) > 1 {
		for _, p := range ports[1:] {
			fmt.Printf("starting server on port: %s\n", p)
			go defaultServer(p)
		}
	}

	fmt.Printf("starting server on port: %s\n", ports[0])
	defaultServer(ports[0])
}
