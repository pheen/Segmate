package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
)

const pageSocketPath = "/tmp/segmate_page.sock"

func startServer() {
	file, err := os.OpenFile("transport.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Set output of the log package to the file
	log.SetOutput(file)

	// Remove the socket file if it already exists
	if _, err := os.Stat(pageSocketPath); err == nil {
		os.Remove(pageSocketPath)
	}

	listener, err := net.Listen("unix", pageSocketPath)
	if err != nil {
		log.Printf("Error starting server: %s\n", err)
		return
	}
	defer listener.Close()

	log.Println("Server is listening on", pageSocketPath)

	for {
		page_conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s\n", err)
			continue
		}

		go handleConnection(page_conn)
	}
}

func handleConnection(page_conn net.Conn) {
	d := json.NewDecoder(page_conn)
	for {
		var m Page // use whatever type is appropriate
		err := d.Decode(&m)
		if err == io.EOF {
			break
		} else if err != nil {
			// handle error
			log.Printf("Error decoding JSON for Page")
		}

		// do something with m
		log.Println(m)
	}
}
