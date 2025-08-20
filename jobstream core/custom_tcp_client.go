package CustomTCPClient

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// create client struct
type Client struct {
	ServerAddr string
	Conn       net.Conn
	Send       chan []byte
}

// create new client
func createClient(ServerAddr string) *Client {
	return &Client{
		ServerAddr: ServerAddr,
		Send:       make(chan []byte, 256),
	}
}

// open client connection
func (client *Client) startTCPConnection() error {
	conn, err := net.Dial("tcp", client.ServerAddr)

	//return if there's error
	if err != nil {
		fmt.Printf("Could not connect to server: %v", err)
		return err
	}

	//register the connection
	client.Conn = conn

	//now write to server
	go client.handleRead()
	go client.handleWrite()

	return nil

}

// handle read
func (client *Client) handleRead() {
	reader := bufio.NewReader(client.Conn)

	//open loop
	for {
		//get the length of new message
		byteLength := make([]byte, 4)
		//get the length
		_, err := io.ReadFull(reader, byteLength)
		if err != nil {
			fmt.Printf("Read error: %v", err)
		}

		length := binary.BigEndian.Uint32(byteLength)
		//get the length of the message
		msg := make([]byte, length)

		_, err = io.ReadFull(reader, msg)
		if err != nil {
			fmt.Println("Read message error:", err)
			return
		}

		fmt.Printf("Received from server: %s\n", string(msg))
	}
}

// handle write
func (client *Client) handleWrite() {
	for msg := range client.Send {
		length := uint32(len(msg))
		buf := make([]byte, 4+len(msg))
		binary.BigEndian.PutUint32(buf[:4], length)
		copy(buf[4:], msg)

		_, err := client.Conn.Write(buf)
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}
	}
}

// close connection
// Close connection
func (c *Client) Close() {
	c.Conn.Close()
	close(c.Send)
}
