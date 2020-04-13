package internal

import (
	"bufio"
	"context"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

// Connect connects to a server via WebSocket.
func Connect(ctx context.Context, serverURL string) (*websocket.Conn, error) {

	wsURL, err := url.Parse(serverURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	query := url.Values{
		"channel": []string{getString(ctx, CtxKeyChannel)},
	}
	wsURL.RawQuery = query.Encode()

	log.Println("connect to:", wsURL.String())
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		// network error
		if resp == nil {
			log.Println(err)
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer resp.Body.Close()

		log.Printf("connection error: %s", data)
		return nil, err
	}
	return conn, nil
}

// ReadFromUserInput ...
func ReadFromUserInput(input chan string) {

	defer close(input)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
		return
	}
}

// ReadFromWebSocket ...
func ReadFromWebSocket(done chan struct{}, conn *websocket.Conn) {

	defer close(done)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("recv: %s", msg)
	}
}

type (
	// KeyChannel ...
	KeyChannel string
)

var (
	// CtxKeyChannel ...
	CtxKeyChannel KeyChannel
)

func getString(ctx context.Context, key interface{}) string {
	// TODO
	val := ctx.Value(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}
