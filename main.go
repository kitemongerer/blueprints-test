package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

var randomWebhookStatusCode = os.Getenv("RANDOM_WEBHOOK_STATUS_CODE")

func main() {
	println("****************************")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("CURL_URL") != "" {
		go pollURL(os.Getenv("CURL_URL"))
	}

	if os.Getenv("LOG_SPAM") != "" {
		println("starting log spam")
		d, err := time.ParseDuration(os.Getenv("LOG_SPAM"))
		if err != nil {
			d = time.Second
		}
		go spamLogs(d)
	}

	if os.Getenv("SLOW_HEALTHCHECK") != "" {
		println("starting with slow healthcheck")
		slowHealthcheck(port, os.Getenv("SLOW_HEALTHCHECK"))
	} else if os.Getenv("PORT_DETECTOR_TEST") == "2" {
		println("starting port detector test 2")
		portDetectorTest2(port)
	} else if os.Getenv("PORT_DETECTOR_TEST") == "3" {
		println("starting port detector test 3")
		portDetectorTest3(port)
	} else if os.Getenv("PORT_DETECTOR_TEST") == "4" {
		println("starting port detector test ephemeral ports")
		portDetectorTestEphemeralPorts()
	} else if os.Getenv("PORT_DETECTOR_TEST") == "5" {
		println("starting port detector test listening on TCP pausing, and then HTTP")
		listenWaitThenServe(port)
	} else if os.Getenv("PORT_DETECTOR_TEST") != "" {
		println("starting default server, secondary server, and udp server")
		portDetectorTest()
	} else if os.Getenv("PORTS") != "" {
		ports := os.Getenv("PORTS")
		println("starting for ports: " + ports)
		startPorts(ports)
	} else if os.Getenv("BIND_ADDR") != "" {
		println("starting with bind addr")
		serveAtAddr(os.Getenv("BIND_ADDR"))
	} else if os.Getenv("ALL_INTERFACES") != "" {
		println("starting on each interface")
		serveInterfaces()
	} else if os.Getenv("PANIC_EVERY") != "" {
		go panicEvery(os.Getenv("PANIC_EVERY"))
		defaultServer(port)
	} else {
		println("starting with default server")
		defaultServer(port)
	}
}

func panicEvery(getenv string) {
	dur, err := time.ParseDuration(getenv)
	if err != nil {
		dur = time.Minute
	}

	time.Sleep(dur)
	panic("time to panic")
}

func spamLogs(d time.Duration) {
	var i int
	for {
		fmt.Printf("LOG NUMBER %d\n", i)
		i++
		time.Sleep(d)
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
	waitForExit()
}

func waitForExit() {
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

	server := defaultHTTPServer(addr)

	log.Println(server.ListenAndServe())
}

func defaultHTTPServer(addr string) *http.Server {
	var server *http.Server
	server = &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("received request at %s\n", r.URL.Path)

		if strings.Contains(r.URL.Path, "webhook") {
			body, _ := io.ReadAll(r.Body)
			fmt.Printf("received webhook body: %s\n", string(body))

			statusCode := http.StatusOK
			if randomWebhookStatusCode != "" {
				statusCode = 100 + rand.Intn(400)
			}

			fmt.Printf("responding with status code: %d\n", statusCode)

			w.WriteHeader(statusCode)
			w.Write([]byte(`{"response": "body"}`))
			return
		}

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
		w.Write([]byte("hi from " + os.Getenv("RENDER_SERVICE_NAME") + server.Addr))
	})}
	return server
}

func listenWaitThenServe(port string) {
	server := defaultHTTPServer(":" + port)
	l := startTCP(port)
	time.Sleep(30 * time.Second)
	log.Println(server.Serve(l))
}

func slowHealthcheck(port string, duration string) {
	dur, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	err = http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%f seconds request (%s) at %s\n", time.Since(start).Seconds(), r.URL.Path, time.Now().String())

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

	udp := startUDP("0")
	defer udp.Close()

	tcp := startTCP("0")
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

func portDetectorTestEphemeralPorts() {
	go defaultServer("8082")
	time.Sleep(5 * time.Second)

	for {
		udp := startUDP("0")
		tcp := startTCP("0")

		time.Sleep(30 * time.Second)
		udp.Close()
		tcp.Close()
	}

}

func startUDP(port string) *net.UDPConn {
	s, err := net.ResolveUDPAddr("udp6", ":"+port)
	if err != nil {
		log.Fatalf("error resolving addr on udp: %s", err)
	}

	conn, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Fatalf("error listening on udp: %s", err)
	}
	return conn
}

func startTCP(port string) *net.TCPListener {
	s, err := net.ResolveTCPAddr("tcp", ":"+port)
	if err != nil {
		log.Fatalf("error resolving addr on tcp: %s", err)
	}

	conn, err := net.ListenTCP("tcp", s)
	if err != nil {
		log.Fatalf("error listening on tcp: %s", err)
	}
	return conn
}

func startPorts(portsList string) {
	ports := strings.Split(portsList, ",")

	for _, p := range ports {
		protocol := "http"
		port := p

		if strings.Contains(p, ":") {
			parts := strings.Split(p, ":")
			if len(parts) != 2 {
				log.Fatalf("invalid port: %s", p)
			}

			protocol = strings.ToLower(parts[0])
			port = parts[1]
		}

		switch protocol {
		case "http":
			fmt.Printf("starting HTTP server on port: %s\n", port)
			go defaultServer(port)
		case "tcp":
			fmt.Printf("starting TCP server on port: %s\n", port)
			go startTCP(port)
		case "udp":
			fmt.Printf("starting UDP server on port: %s\n", port)
			go startUDP(port)
		default:
			log.Fatalf("invalid protocol: %s", protocol)
		}
	}

	waitForExit()
}
