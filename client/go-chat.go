package main

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//go:embed assets/noti.mp3
var soundAssets embed.FS

func play_sound(sound string) {
	f, err := soundAssets.Open(sound)
	if err != nil {
		log.Fatal("sound open:", err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal("mp3 decode:", err)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(beep.Seq(
		streamer,
		beep.Callback(func() {
			streamer.Close()
		}),
	))

}

func main() {
	var display_name string

	fmt.Printf("\nEnter your display name: ")
	fmt.Scan(&display_name)

	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println(err)
	}

	//Send display name immediately
	_, err = conn.Write([]byte(display_name + "\n"))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Print("Connected to Server!\n\n")

	go func() {
		rd := bufio.NewReader(conn)
		for {
			raw, err := rd.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("<server disconnected>")
				} else {
					fmt.Println("<read error>")
				}
				log.Fatal(err)
			}

			line := strings.TrimRight(raw, "\r\n")

			parts := strings.SplitN(line, ": ", 2)
			if len(parts) != 2 {
				fmt.Printf("\r%s\n%s: ", line, display_name)
				continue
			}
			sender, body := parts[0], parts[1]

			if sender == display_name {
				continue
			}

			fmt.Printf("\r%s: %s\n%s: ", sender, body, display_name)
			play_sound("assets/noti.mp3")

		}
	}()

	for {
		stdin := bufio.NewReader(os.Stdin)

		fmt.Printf("%s: ", display_name)
		msg, err := stdin.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading user input: %q", err)
		}

		msg = strings.TrimRight(msg, "\r\n")
		_, err = conn.Write([]byte(display_name + ": " + msg + "\n"))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
