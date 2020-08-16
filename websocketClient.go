// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/connection/test"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	requestPayload := map[string]interface{}{}

	go func() {
		for {
			log.Println("reading payload")
			err := c.ReadJSON(&requestPayload)
			if err != nil {
				log.Println("read:", err)
				c.WriteJSON(map[string]string{"error": err.Error()})
				continue
			}
			log.Println("payload read completely")
			log.Println(requestPayload)

			err = c.WriteJSON(map[string]interface{}{"body": "hello, recieved you payload successfully", "received_payload": requestPayload["body"], "recieved_url": requestPayload["url"]})
			log.Println("error while writing data into connection", err)
		}
	}()

	for {

		select {
		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		}

	}

}
