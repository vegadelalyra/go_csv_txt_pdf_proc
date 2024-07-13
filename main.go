package main

import (
	"bufio"
	"log"
	"os"
)

func main() {
	// Open the file
	// Method 1: Using os.Readfile
	log.Println("--- os.ReadFile ---")
	content, err := os.ReadFile("data/example.txt")
	if err != nil {
		log.Fatal(err)
	}
	// Do something
	// Parse/Process
	log.Println(string(content))

	// Method 2: Using os.Open
	log.Println("--- os.Open ---")
	f, err := os.Open("data/example.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Parser/Process
	// Create A Scanner
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		log.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
