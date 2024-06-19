package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func onMessage(conn net.Conn) {
	for {
		reader := bufio.NewReader(conn)
		msg, _ := reader.ReadString('\n')
		fmt.Print(msg)
	}
}

func main() {
	connection, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter your Name!: ")
	nameReader := bufio.NewReader(os.Stdin)
	nameInput, _ := nameReader.ReadString('\n')
	nameInput = strings.TrimSpace(nameInput)

	fmt.Println("********** MESSAGES **********")

	go onMessage(connection)

	for {
		msgReader := bufio.NewReader(os.Stdin)
		msg, err := msgReader.ReadString('\n')
		if err != nil {
			break
		}

		msg = strings.TrimSpace(msg)
		if strings.HasPrefix(msg, "/create-channel ") || strings.HasPrefix(msg, "/join ") {
			connection.Write([]byte(msg + "\n"))
		} else {
			msg = fmt.Sprintf("%s: %s", nameInput, msg)
			connection.Write([]byte(msg + "\n"))
		}
	}

	connection.Close()
}
