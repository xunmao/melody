package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var (
	channel = flag.String("channel", "test", "Specify a channel.")
)

func main() {

	flag.Parse()

	wsURL, err := url.Parse("ws://localhost:8080/messaging")
	if err != nil {
		log.Fatalln(err)
	}
	query := url.Values{
		"channel": []string{*channel},
	}
	wsURL.RawQuery = query.Encode()

	log.Println("connect to:", wsURL.String())
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		if resp == nil {
			log.Fatalln(err)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
		log.Fatalln(string(data))
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("recv: %s", msg)
		}
	}()

	send := make(chan string)
	go func() {
		defer close(send)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			send <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Println(err)
			return
		}
	}()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-done:
			return
		case msg := <-send:
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
