package main

import (
	"crypto/tls"
	"dero-swap/cfg"
	"dero-swap/coin"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var Connection *websocket.Conn

func StartClient(server url.URL) {

	go func() {
		ticker := time.NewTicker(time.Minute * 2)
		defer ticker.Stop()
		for range ticker.C {
			UpdatePool()
		}
	}()

	for {

		var incoming WS_Message

		dialer := websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, _, err := websocket.DefaultDialer.Dial(server.String(), nil)
		if err != nil {
			log.Println("Websocket error, re-connect in 10 seconds:", err)
			time.Sleep(time.Second * 10)
			continue
		}
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		Connection = conn
		go func(c *websocket.Conn) {
			ticker := time.NewTicker(time.Second * 30)
			defer ticker.Stop()
			for range ticker.C {
				if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
					log.Println("SetReadDeadline error:", err)
					Connection.Close()
					return
				}
				if err := Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("Ping error:", err)
					Connection.Close()
					return
				}
			}
		}(conn)

		UpdatePool()
		log.Printf("Connected to server %s\n", cfg.Settings.ServerAddress)

		for {
			if err := Connection.ReadJSON(&incoming); err != nil {
				log.Println("Websocket error, re-connect in 30 seconds:", err)
				break
			}

			var outgoing WS_Message
			outgoing.ID = incoming.ID

			switch incoming.Method {
			case "swap":

				var request coin.Swap_Request
				var response coin.Swap_Response

				p := incoming.Params.(map[string]any)
				outgoing.Method = "client_ok"

				if d, ok := p["pair"].(string); ok {
					request.Pair = d
				}
				if d, ok := p["amount"].(float64); ok {
					request.Amount = d
				}
				if d, ok := p["address"].(string); ok {
					request.Address = d
				}
				if d, ok := p["price"].(float64); ok {
					request.Price = d
				}

				isReverse := false
				if strings.HasPrefix(request.Pair, "dero") {
					isReverse = true
				}

				if !isReverse {
					response = Dero_Swap(request)
				} else {
					response = Reverse_Swap(request)
				}

				outgoing.Result = response

				Connection.WriteJSON(outgoing)
			}
		}
		time.Sleep(time.Second * 30)
	}
}
