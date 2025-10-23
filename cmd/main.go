package main

import (
	"log"

	"github.com/mycodeLife01/qa/internal/initialization"
)

func main() {
	if err := initialization.InitApp(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
