package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	conns = &sync.Map{}
	names = &sync.Map{}
)

var chatContext []openai.ChatCompletionMessageParamUnion

var api_key string

func chat(api_key string) (rsp string) {
	client := openai.NewClient(
		option.WithAPIKey(api_key),
	)
	ctx := context.Background()

	// Needed to include instruction in the system message to not include newlines in the reponse to prevent trimming of the rendered message in chat ui
	chatContext = append(chatContext, openai.SystemMessage("you are a gen z kid in a groupchat. use gen z slang and typeface. DO NOT USE NEWLINES IN YOUR RESPONSE."))

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: chatContext,
		Seed:     openai.Int(0),
		Model:    openai.ChatModelGPT4_1Mini,
	})

	if err != nil {
		fmt.Println("Error sending request to OpenAI API: ", err)
		return ""
	}

	rsp = completion.Choices[0].Message.Content

	fmt.Println("CHAT RESPONSE: ", rsp)

	return rsp
}

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

		//TODO find out why this fixes the issue where only the sender of the room command receives the result
		time.Sleep(1 * time.Second)

		//Find command in user message, server sends message
		t, command := findCommand(strings.Trim(string(buffer[:n]), "\n"))

		//debug
		fmt.Printf("<findCommand : %t | %s>\n", t, command)

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
					command_return := fmt.Sprintf("Connected Users %v\n", "["+users+"]")
					broadcastMsg(nil, conns, command_return)
				}()
			case "chat":
				go func() {
					if api_key != "" {
						b, prompt := findPrompt(strings.Trim(string(buffer[:n]), "\n"))
						prompt = display_name + ": " + prompt
						if b {
							chatContext = append(chatContext, openai.UserMessage(prompt))
							//fmt.Println("Sending prompt: ", prompt)
							raw_rsp := chat(api_key)
							rsp := fmt.Sprintf("AI: %s\n", raw_rsp)
							chatContext = append(chatContext, openai.AssistantMessage(raw_rsp))
							broadcastMsg(nil, conns, rsp)
						}
					} else {
						fmt.Println("<No API Key Found>")
						broadcastMsg(nil, conns, "<No API Key Found>\n")
					}
				}()
			default:
				continue
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

// pretty ugly, need to change this and use regex
func findCommand(msg string) (bool, string) {
	cmd_index := strings.Index(msg, "#")

	if cmd_index != -1 {
		space_index := strings.Index(msg[cmd_index+1:], " ")

		if space_index != -1 {
			return true, msg[cmd_index+1:][:space_index]
		}

	} else {
		return false, ""
	}

	return true, msg[cmd_index+1:]

}

func findPrompt(msg string) (bool, string) {
	open_quote := strings.Index(msg, `"`)

	if open_quote != -1 {
		close_quote := strings.Index(msg[open_quote+1:], `"`)

		if close_quote != -1 {
			return true, msg[open_quote+1:][:close_quote]
		} else {
			return false, ""
		}
	} else {
		return false, ""
	}
}

func main() {
	godotenv.Load(".env")
	api_key = os.Getenv("OPENAI_API_KEY")

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
