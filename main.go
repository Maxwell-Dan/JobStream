package main

import (
	"fmt"
	"time"
   "github.com/Maxwell-Dan/JobStream/CustomTCPServer"
)

func main() {
	// Start the server
	go func() {
		server := CustomTCPServer.NewServer(":9000")
		if err := CustomTCPServer.StartTCPConnection(server); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	time.Sleep(time.Second) // Give server time to start

	// Start the client
	client := customtcpclient.NewClient(":9000")
	if err := client.Connect(); err != nil {
		fmt.Println("Client connection error:", err)
		return
	}

	// Send a message from client to server
	client.Send <- []byte("Hello from client!")

	time.Sleep(2 * time.Second)
	client.Close()
}
