package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var conMap = &sync.Map{}

func handleConnections(conn net.Conn) {
	defer func() {
		conn.Close()
		conMap.Delete(conn.RemoteAddr().String())
		broadcastMsg(nil, conMap, fmt.Sprintf("<%s left the chat>\n", conn.RemoteAddr().String()))
		fmt.Printf("\n%s left the chat\n", conn.RemoteAddr().String())
	}()
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("%s | %s\n", strings.Trim(string(buffer[:n]), "\n"), conn.RemoteAddr().String())
		broadcastMsg(conn, conMap, string(buffer[:n]))
	}
}

func broadcastMsg(sender net.Conn, conMap *sync.Map, msg string) {
	conMap.Range(func(key any, value any) bool {
		client := value.(net.Conn)

		if sender == client {
			return true
		}

		_, err := client.Write([]byte(msg))
		if err != nil {
			fmt.Println(err)
		}
		return true
	})
}

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nServer listening on Port 8000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		conMap.Store(conn.RemoteAddr().String(), conn)
		fmt.Printf("\nNew Connection: %s\n\n", conn.RemoteAddr().String())
		broadcastMsg(conn, conMap, fmt.Sprintf("<%s joined the chat>\n", conn.RemoteAddr().String()))
		go handleConnections(conn)
	}

}
