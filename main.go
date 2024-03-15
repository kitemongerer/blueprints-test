package main

import (
	"time"
)

func main() {
	for {
		println("still here...")
		time.Sleep(5 * time.Second)
	}
}
