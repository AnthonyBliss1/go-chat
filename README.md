# Go Chat

A chat room style project utilizing a TCP server (`server.go`) to broadcast messages to connected clients. Use `go-chat.go` to establish a connection to the server and send messages. 

The client side code, `go-chat.go`, will prompt the user for a server address to connect to. Also, `server.go` is currently set up to listen on `Port 8000` so make sure this port is not in use or change the port configuration in `go-chat.go` and `server.go`.

## Commands
***Send commands with `#`***

| Command | Usage |
| ------- | ----- |
| #room | Show the current users connected to the server |
| #chat "{prompt}" | Send a message to the AI bot |

- To the use the `#chat` command, you need a `.env` file that includes your `OPENAI_API_KEY` next to the server code.
- Make sure to wrap your prompt in quotes:
    - `#chat "Hello!"`

## Usage 

1. **Clone the Repo**
```bash
git clone https://github.com/AnthonyBliss1/go-chat.git
```

2. **Run Server and Client Packages** 
```bash 
go run server/server.go
```

```bash 
go run app/go-chat.go
```

3. **(Optional) Build Executable**
```bash 
go build -o builds/server ./server/server.go
go build -o builds/go-chat ./client/go-chat.go
```

4. **(Optional) Run Executable**
```bash 
./builds/server
```

```bash
./builds/go-chat
```