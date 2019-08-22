package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	"github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
	"github.com/sirupsen/logrus"
)

var (
	take     = flag.Int("take", 0, "take only the first n trades. If unset, continue until interrupt.")
	logLevel = flag.Int("log-level", int(logrus.InfoLevel), "set numeric log level per https://godoc.org/github.com/sirupsen/logrus#Level")
)

func check(err error, context string) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// SubscribeTrade generates a request to subscribe to a trade feed
func SubscribeTrade(symbol string, precision int) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"subscribe": "trade",
		"symbol":    symbol,
		"precision": precision,
	})
}

func listen(conn *websocket.Conn, messages chan<- []byte) {
	defer close(messages)

	for {
		mtype, message, err := conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Error("failed to read message")
			return
		}
		switch mtype {
		case websocket.CloseMessage:
			logrus.WithField("data", message).Info("received close messagse from server")
			return
		case websocket.TextMessage, websocket.BinaryMessage:
			messages <- message
		case websocket.PongMessage:
			// ignore it; it's pretty unlikely without our originating pings, anyway
		case websocket.PingMessage:
			logrus.Debug("ignoring ping message originating from server")
		}

	}
}

func prettyJSON(bytes []byte) (s string, err error) {
	var obj interface{}
	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		return
	}
	bytes, err = json.MarshalIndent(obj, "", "  ")
	s = string(bytes)
	return
}

func main() {
	flag.Parse()

	logrus.SetLevel(logrus.Level(*logLevel))

	conn, _, err := websocket.DefaultDialer.Dial(bitmart.WSSBitmart, nil)
	check(err, "dial")
	defer conn.Close()

	st, err := SubscribeTrade(bitmart.NdauSymbol, 4)
	check(err, "make subscribe json")
	logrus.WithField("subscribe", string(st)).Debug("subscribe to XND/USDT")

	err = conn.WriteMessage(websocket.TextMessage, st)
	check(err, "send")

	messages := make(chan []byte)
	go listen(conn, messages)

	for {
		select {
		case message, ok := <-messages:
			if !ok {
				return
			}
			js, err := prettyJSON(message)
			if err != nil {
				logrus.WithError(err).Error("failed to prettify JSON")
				fmt.Printf("%s\n", message)
			} else {
				fmt.Println(js)
			}
		}
	}
}
