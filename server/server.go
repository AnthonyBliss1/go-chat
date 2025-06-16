package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
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
		broadcastMsg(nil, conns, fmt.Sprintf("<%s left the room>\n", display_name))
		fmt.Printf("\n%s | %s left the room\n", display_name, conn.RemoteAddr().String())
	}()

	names.Store(id, display_name)
	conns.Store(conn.RemoteAddr().String(), conn)
	fmt.Printf("\nNew Connection: %s | %s\n\n", display_name, conn.RemoteAddr().String())
	broadcastMsg(conn, conns, fmt.Sprintf("<%s joined the room>\n", display_name))

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("%s | %s\n", strings.Trim(string(buffer[:n]), "\n"), conn.RemoteAddr().String())
		broadcastMsg(conn, conns, string(buffer[:n]))

		//TODO find out why this fixes the issue where only the sender of the command receives the result 😐
		time.Sleep(1 * time.Second)

		//Find command in user message, server sends message
		t, command := findCommand(strings.Trim(string(buffer[:n]), "\n"))
		if t {
			switch command {
			case "room":
				go func() {
					var list []string
					names.Range(func(_, value any) bool {
						list = append(list, value.(string))
						return true
					})
					users := strings.Join(list, ", ")
					command_return := fmt.Sprintf("Connected Users: %v\n", "["+users+"]")
					broadcastMsg(nil, conns, command_return)
				}()
			default:
				return
			}
		}
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

func findCommand(msg string) (bool, string) {
	index := strings.Index(msg, "#")
	if index == -1 {
		return false, ""
	}

	command_string := strings.TrimSpace(msg[index+1:])
	if command_string == "" {
		return false, ""
	}

	return true, command_string

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
