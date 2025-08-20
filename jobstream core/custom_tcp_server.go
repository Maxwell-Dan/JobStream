package CustomTCPServer

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

//a TCP Connection is a consistent connection between two nodes
// that is the client and server
//to easily handle the properties of the client and server you have to define a struct to make things easy
// the struct of the client includes the IP of the client, or the ID, then it has the connection
//and also the channel the client and send and receive message to the server.
//channel is used to controll the read and writes of multiple clients to the client using concurrency

// struct of the client which consists the client ID, Conn, and Chan [] bye
type Client struct {
	ID   string
	Conn net.Conn
	Send chan []byte
}

//the server receives the connection of the client, because it can receive
//  multiple clients it should be contained in a map
// it has to control the read and write of the clients, of multiple clients to prevent it from crashing

type Server struct {
	Addr    string
	Clients map[string]*Client
	Mu      sync.Mutex
}

// create new server and return it
func newServer(addr string) *Server {
	//return the server object
	return &Server{
		Addr:    addr,
		Clients: make(map[string]*Client),
	}
}

// now that a new server has been created, you can now start with the connection
func startTCPConnection(server *Server) error {

	//listen to the server port from the server struct
	listener, err := net.Listen("tcp", server.Addr)

	if err != nil {
		return err
	}

	fmt.Println("Server started on", server.Addr)

	//now that connection has started, you have to defer the close of the connection once it ends
	defer listener.Close()

	//now you can loop through to accept connections
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("Something wrong happend: %v", err)
		}

		//now that I have the client address, I can now get its address and use it in the server
		clientID := conn.RemoteAddr().String()
		//access the client and register the properties
		client := &Client{
			ID:   clientID,
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		//okay now I can register the client details in the server,
		// but I have to ensure the server is not overwhelmed
		server.Mu.Lock()
		server.Clients[clientID] = client
		server.Mu.Unlock()

		//read and write using channels
		go server.handleRead(client)
		go server.handleWrite(client)
	}
}

// create the read
func (server *Server) handleRead(client *Client) {

	defer func() {
		server.removeClient(client)
	}()

	//read client's message
	reader := bufio.NewReader(client.Conn)

	//the message is encoded in bytes, and it has to be translated to a string
	// and there has to be a loop so that it is real time
	for {
		//the messagees are sent in bytes so determine the length of the byte
		byteLength := make([]byte, 4)

		//read the client message with the length of the byte as a guide
		_, err := io.ReadFull(reader, byteLength)

		if err != nil {
			fmt.Printf("Could not read the full length %v", err)
			return
		}

		//encode the byte into unsigned interger
		length := binary.BigEndian.Uint32(byteLength)

		//create the size of the message
		msg := make([]byte, length)

		//read the message
		_, err = io.ReadFull(reader, msg)

		if (err) != nil {
			fmt.Println("Something went wrong: ", err)
			return
		}

		fmt.Printf("Received from %s: %s\n", client.ID, string(msg))

	}

}

// create the write
func (s *Server) handleWrite(c *Client) {
	defer func() {
		c.Conn.Close()
	}()

	for msg := range c.Send {
		length := uint32(len(msg))
		buf := make([]byte, 4+len(msg))
		binary.BigEndian.PutUint32(buf[:4], length)
		copy(buf[4:], msg)

		_, err := c.Conn.Write(buf)
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}
	}
}

//clean client

func (s *Server) removeClient(c *Client) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	fmt.Println("Client disconnected:", c.ID)
	c.Conn.Close()
	delete(s.Clients, c.ID)
	close(c.Send)
}
