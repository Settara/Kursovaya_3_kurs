package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

var clients = make(map[string]clientInfo) // Tracks connected clients (username, connection)
var mu sync.Mutex                         // Ensures concurrent access safety for clients map

type clientInfo struct {
	username string
	conn     net.Conn
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()

	// Request and validate the password
	conn.Write([]byte("Enter server password: "))
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		password := scanner.Text()
		if password != ServerPassword {
			conn.Write([]byte("Incorrect password. Connection closed.\n"))
			log.Printf("Client %s failed to provide correct password.\n", clientAddr)
			return
		}
	}

	// Request username
	conn.Write([]byte("Enter your username: "))
	var username string
	if scanner.Scan() {
		username = scanner.Text()
	}

	log.Printf("Client connected: %s (Username: %s)\n", clientAddr, username)
	fmt.Printf("New client connected: %s (Username: %s)\n", clientAddr, username)

	// Store the client's information
	mu.Lock()
	clients[clientAddr] = clientInfo{username: username, conn: conn}
	mu.Unlock()

	for scanner.Scan() {
		message := scanner.Text()
		switch {
		case strings.HasPrefix(message, "/msg"):
			handlePrivateMessage(clientAddr, message)
		case message == "/list":
			handleListClients(conn)
		default:
			log.Printf("Unknown command from %s: %s", clientAddr, message)
			conn.Write([]byte("Unknown command. Use /help for a list of available commands.\n"))
		}
	}

	// Client disconnect handling
	mu.Lock()
	delete(clients, clientAddr)
	mu.Unlock()
	log.Printf("Client disconnected: %s (Username: %s)\n", clientAddr, username)
	fmt.Printf("Client disconnected: %s (Username: %s)\n", clientAddr, username)
}

func handlePrivateMessage(senderAddr, message string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		log.Printf("Invalid message format from %s: %s", senderAddr, message)
		return
	}
	recipientAddr := parts[1]
	msgText := parts[2]

	mu.Lock()
	recipientInfo, exists := clients[recipientAddr]
	mu.Unlock()

	if exists {
		_, err := recipientInfo.conn.Write([]byte(fmt.Sprintf("Message from %s: %s\n", clients[senderAddr].username, msgText)))
		if err != nil {
			log.Printf("Error sending message to %s: %v", recipientAddr, err)
		} else {
			log.Printf("Message sent from %s to %s: %s", clients[senderAddr].username, recipientInfo.username, msgText)
		}
	} else {
		log.Printf("Recipient %s not found", recipientAddr)
		senderConn := clients[senderAddr].conn
		senderConn.Write([]byte(fmt.Sprintf("Recipient %s not found. Please check the address.\n", recipientAddr)))
	}
}

// Enhanced "/list" command to show usernames and IP addresses
func handleListClients(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	clientList := "Connected clients:\n"
	for addr, info := range clients {
		clientList += fmt.Sprintf("Username: %s, Address: %s\n", info.username, addr)
	}
	conn.Write([]byte(clientList))
}

func startServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server started on port %s\n", port)
	fmt.Printf("Server started on port %s. Waiting for clients...\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleClient(conn) // Handle each connection in a separate goroutine
	}
}

func main() {
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	port := "8080" // Default port
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	startServer(port)
}
