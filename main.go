package main

import (
	"github.com/kirugan/aviasales/proxy"
	"time"
)

const defaultTimeout = 3 * time.Second

func main() {
	proxy.Start("8080", defaultTimeout)
}
