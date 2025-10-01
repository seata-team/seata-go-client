package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <example>")
		fmt.Println("Available examples:")
		fmt.Println("  basic        - Basic client usage")
		fmt.Println("  saga         - Saga pattern example")
		fmt.Println("  tcc          - TCC pattern example")
		fmt.Println("  comprehensive - Comprehensive example with all features")
		return
	}

	example := os.Args[1]

	switch example {
	case "basic":
		basicExample()
	case "saga":
		sagaExample()
	case "tcc":
		tccExample()
	case "comprehensive":
		comprehensiveExample()
	default:
		log.Fatalf("Unknown example: %s", example)
	}
}
