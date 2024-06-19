package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Channel struct {
	name  string
	conns []net.Conn
	mu    sync.Mutex
}

var (
	channels = make(map[string]*Channel)
	connCh   = make(chan net.Conn)
	closeCh  = make(chan net.Conn)
	msgCh    = make(chan Message)
)

type Message struct {
	conn    net.Conn
	channel *Channel
	text    string
}

func main() {
	server, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				log.Fatal(err)
			}
			connCh <- conn
			fmt.Println("Someone joined the server!")
		}
	}()

	for {
		select {
		case conn := <-connCh:
			go onMessage(conn)

		case msg := <-msgCh:
			fmt.Println(msg.text)
			publishMsg(msg)

		case conn := <-closeCh:
			fmt.Println("client exit")
			removeConn(conn)
		}
	}
}

func removeConn(conn net.Conn) {
	for _, channel := range channels {
		channel.mu.Lock()
		for i, c := range channel.conns {
			if c == conn {
				channel.conns = append(channel.conns[:i], channel.conns[i+1:]...)
				break
			}
		}
		channel.mu.Unlock()
	}
}

func publishMsg(msg Message) {
	msg.channel.mu.Lock()
	defer msg.channel.mu.Unlock()
	fmt.Printf("Broadcasting message to channel: %s\n", msg.channel.name)
	for _, conn := range msg.channel.conns {
		if conn != msg.conn {
			_, err := conn.Write([]byte(msg.text + "\n"))
			if err != nil {
				fmt.Printf("Error sending message to a client: %v\n", err)
			}
		}
	}
}

func onMessage(conn net.Conn) {
	reader := bufio.NewReader(conn)
	var currentChannel *Channel

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		msg = strings.TrimSpace(msg)
		if strings.HasPrefix(msg, "/create-channel ") {
			channelName := strings.TrimPrefix(msg, "/create-channel ")
			if _, exists := channels[channelName]; !exists {
				channels[channelName] = &Channel{name: channelName}
				currentChannel = channels[channelName]
				currentChannel.conns = append(currentChannel.conns, conn)
				conn.Write([]byte(fmt.Sprintf("Channel '%s' created and joined.\n", channelName)))
			} else {
				conn.Write([]byte("Channel already exists.\n"))
			}
		} else if strings.HasPrefix(msg, "/join ") {
			channelName := strings.TrimPrefix(msg, "/join ")
			if channel, exists := channels[channelName]; exists {
				currentChannel = channel
				currentChannel.mu.Lock()
				currentChannel.conns = append(currentChannel.conns, conn)
				currentChannel.mu.Unlock()
				conn.Write([]byte(fmt.Sprintf("Joined channel '%s'.\n", channelName)))
				fmt.Printf("Joined channel '%s'.\n", channelName)
			} else {
				conn.Write([]byte("Channel does not exist.\n"))
			}
		} else {
			if currentChannel != nil {
				msgCh <- Message{conn: conn, channel: currentChannel, text: msg}
			} else {
				conn.Write([]byte("You need to join a channel first. Use /join <channel> or /create-channel <channel>.\n"))
			}
		}
	}

	closeCh <- conn
}
