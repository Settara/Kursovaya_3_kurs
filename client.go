package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func printHelp() {
	helpMessage := `
Available commands:
- /msg : Send a private message to another client (you'll specify the recipient next).
- /list : View the list of all connected clients (usernames and IPs).
- /help : Show available commands.
- /exit : Exit the messenger.
`
	fmt.Println(helpMessage)
}

func startClient(serverAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Handle incoming messages from the server in a separate goroutine
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Printf("\n%s\n", scanner.Text()) // Ensure new messages start from a new line
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	// Request password
	fmt.Print("Enter server password: ")
	if scanner.Scan() {
		password := scanner.Text()
		_, err := conn.Write([]byte(password + "\n"))
		if err != nil {
			log.Fatalf("Failed to send password: %v", err)
		}
	}

	// Request username
	fmt.Print("Enter your username: ")
	var username string
	if scanner.Scan() {
		username = scanner.Text()
		_, err := conn.Write([]byte(username + "\n"))
		if err != nil {
			log.Fatalf("Failed to send username: %v", err)
		}
	}

	fmt.Println("Connected to the server as", username)
	fmt.Println("Type '/help' for a list of commands.")

	// Main input loop for the user
	for {
		fmt.Print("Enter command or message (or /help for commands, /exit to quit): ")
		if scanner.Scan() {
			command := scanner.Text()

			switch {
			case command == "/exit":
				fmt.Println("Exiting client.")
				return
			case command == "/help":
				printHelp()
			case command == "/list":
				_, err := conn.Write([]byte("/list\n"))
				if err != nil {
					log.Printf("Error sending /list command: %v", err)
				}
			case strings.HasPrefix(command, "/msg"):
				// Ask for the recipient after the message is typed
				fmt.Print("Enter recipient's IP:Port: ")
				if scanner.Scan() {
					recipient := scanner.Text()

					// Ensure the message isn't empty
					message := strings.TrimSpace(command[4:])
					if message == "" {
						fmt.Println("Message cannot be empty.")
					} else {
						// Format and send the message to the server
						fullMessage := fmt.Sprintf("/msg %s %s\n", recipient, message)
						_, err := conn.Write([]byte(fullMessage))
						if err != nil {
							log.Printf("Error sending message: %v", err)
						} else {
							fmt.Println("Message sent.")
						}
					}
				}
			default:
				// For normal messages, prompt for the recipient after the message
				fmt.Print("Enter recipient's IP:Port: ")
				if scanner.Scan() {
					recipient := scanner.Text()

					// Ensure the message isn't empty
					if strings.TrimSpace(command) == "" {
						fmt.Println("Message cannot be empty.")
					} else {
						// Format and send the message to the server
						fullMessage := fmt.Sprintf("/msg %s %s\n", recipient, command)
						_, err := conn.Write([]byte(fullMessage))
						if err != nil {
							log.Printf("Error sending message: %v", err)
						} else {
							fmt.Println("Message sent.")
						}
					}
				}
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <server_address>")
		return
	}

	serverAddr := os.Args[1]
	startClient(serverAddr)
}
