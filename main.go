package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Config struct {
	WriteKey string `json:"writeKey"`
}

const idleTimeout = 5 * time.Second

func init() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startServer()
		os.Exit(0)
	}
}

func main() {
	is_parent_process := !fiber.IsChild()
	if is_parent_process {
		fmt.Printf("Calling server start")
		// Start the server in a separate child process
		cmd := exec.Command(os.Args[0], "server")
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error starting server process: %s\n", err)
			return
		}
	} else {
		// fmt.Println("I'm a child process")
		// fmt.Println(myString)
	}

	// Sources Config
	sourcesConfigFile, err := os.Open("sources_config.json")
	if err != nil {
		log.Fatalf("Error opening config file: %s", err)
	}
	defer sourcesConfigFile.Close()

	var sources_config map[string]interface{}
	if err := json.NewDecoder(sourcesConfigFile).Decode(&sources_config); err != nil {
		log.Fatalf("Error parsing JSON file: %s", err)
	}

	// Transport Conn
	// address := "localhost:8080"
	retryInterval := 100 * time.Millisecond
	maxRetries := 5
	waitForSocket(pageSocketPath, retryInterval, maxRetries)

	page_conn, err := net.Dial("unix", pageSocketPath)
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		return
	}
	defer page_conn.Close()

	app := fiber.New(fiber.Config{
		IdleTimeout: idleTimeout,
		// Prefork: true,
	})

	// page
	app.Post("/v1/p", func(c *fiber.Ctx) error {
		// todo:
		//   - dedup by messageId
		//   - reject if write key doesn't match the expected url?
		//   - probably ensure the message is valid at this level

		// config := new(Config)
		// if err := c.BodyParser(config); err != nil {
		// 	return c.JSON(fiber.Map{"error": "0"})
		// }

		// _, exists := sources_config[config.WriteKey]
		// if !exists {
		// 	return c.JSON(fiber.Map{"error": "1"})
		// }

		// message := "Hello from client"
		// _, err = conn.Write([]byte(message))
		_, err = page_conn.Write(c.BodyRaw())
		if err != nil {
			fmt.Printf("Error writing to server: %s\n", err)
			// return
		}

		fmt.Printf("Success writing to server")

		// log.Println(config.WriteKey)

		// page := new(Page)

		// if err := c.BodyParser(page); err != nil {
		// 	return err
		// }

		return c.JSON(fiber.Map{"error": ""})
	})

	// track
	app.Get("/v1/t", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	// identify
	app.Get("/v1/i", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	// group
	app.Get("/v1/g", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	// screen
	app.Get("/v1/s", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	// alias
	app.Get("/v1/a", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	// batch
	app.Get("/v1/batch", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	// import
	app.Get("/v1/batch", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	// Listen from a different goroutine
	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

	_ = <-c // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	fmt.Println("Running cleanup tasks...")

	// Your cleanup tasks go here
	// db.Close()
	// redisConn.Close()
	fmt.Println("Fiber was successful shutdown.")
}

func waitForSocket(address string, retryInterval time.Duration, maxRetries int) (net.Conn, error) {
	var conn net.Conn
	var err error

	for retries := 0; retries < maxRetries; retries++ {
		conn, err = net.Dial("tcp", address)
		if err == nil {
			return conn, nil // Socket is ready
		}
		fmt.Println("Waiting for socket, retrying...")
		time.Sleep(retryInterval)
	}

	return nil, fmt.Errorf("socket not ready after %d retries: %v", maxRetries, err)
}
