package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xunmao/melody/examples/messaging/client/internal"
)

var (
	channel = flag.String("channel", "test", "Specify a channel.")
)

func main() {

	flag.Parse()

	ctx := context.Background()
	ctx = context.WithValue(ctx, internal.CtxKeyChannel, *channel)

	conn, err := internal.Connect(ctx, "ws://localhost:8080/messaging")
	if err != nil {
		return
	}
	defer conn.Close()

	done := make(chan struct{})
	go internal.ReadFromWebSocket(done, conn)

	input := make(chan string)
	go internal.ReadFromUserInput(input)

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-done:
			return
		case msg := <-input:
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println(err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			err := conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(time.Second),
			)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
