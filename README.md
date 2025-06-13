# Go Chat

A chat room style project utilizing a TCP server (`server.go`) to broadcast messages to connected clients. Use `go-chat.go` to connect to establish a connection to the server and send messages. 

The client side code, `go-chat.go`, is setup to connect to a localhost address @ 127.0.0.1 (although this can be changed). Also, `server.go` is currently set up to listen on `Port 8000`.

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