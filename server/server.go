package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	conns = &sync.Map{}
	names = &sync.Map{}
)

func handleConnections(conn net.Conn) {
	// Read and store connected user display name
	id := fmt.Sprintf("%p", conn)
	rd := bufio.NewReader(conn)

	name_line, err := rd.ReadString('\n')
	if err != nil {
		fmt.Printf("error reading display name: %q\n", err)
		return
	}

	display_name := strings.TrimSpace(name_line)

	defer func() {
		conn.Close()
		conns.Delete(conn.RemoteAddr().String())
		names.Delete(id)
		broadcastMsg(nil, conns, fmt.Sprintf("<%s left the chat>\n", display_name))
		fmt.Printf("\n%s | %s left the chat\n", display_name, conn.RemoteAddr().String())
	}()

	names.Store(id, display_name)
	conns.Store(conn.RemoteAddr().String(), conn)
	fmt.Printf("\nNew Connection: %s | %s\n\n", display_name, conn.RemoteAddr().String())
	broadcastMsg(conn, conns, fmt.Sprintf("<%s joined the chat>\n", display_name))

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("%s | %s\n", strings.Trim(string(buffer[:n]), "\n"), conn.RemoteAddr().String())
		broadcastMsg(conn, conns, string(buffer[:n]))
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

		go handleConnections(conn)
	}

}
