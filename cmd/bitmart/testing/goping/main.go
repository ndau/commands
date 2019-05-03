package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

const bitmart = "wss://openws.bitmart.com"

func check(err error, context string) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(bitmart, nil)
	check(err, "dial")
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"subscribe":"ping"}`))
	check(err, "send")

	_, message, err := conn.ReadMessage()
	check(err, "recv")

	var jsdata interface{}
	err = json.Unmarshal(message, &jsdata)
	check(err, "json unmarshal")
	pretty, err := json.MarshalIndent(jsdata, "", "  ")
	check(err, "json remarshal")
	fmt.Printf("%s\n", pretty)
}
