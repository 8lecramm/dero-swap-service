package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/deroproject/derohe/globals"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"

	"dero-swap/cfg"
	"dero-swap/clients"
	"dero-swap/coin"
	"dero-swap/dero"
)

type WS_Message struct {
	ID     uint64 `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
	Result any    `json:"result,omitempty"`
}

var WSConnections sync.Map

// create handlers for methods
var ws_handlers = make(map[string]func(WS_Message, *websocket.Conn) (WS_Message, error))

// swap other coins to Dero
func Dero_Swap(request coin.Swap_Request) (response coin.Swap_Response) {

	var err error

	// check if destination wallet is valid. Registered usernames can also be used.
	if strings.HasPrefix(request.Address, "dero1") || strings.HasPrefix(request.Address, "deroi") {
		_, err = globals.ParseValidateAddress(request.Address)
	} else {
		if addr := dero.CheckAddress(request.Address); addr != "" {
			request.Address = addr
		} else {
			err = fmt.Errorf("invalid address")
		}
	}

	// basic checks
	if request.Amount == 0 || err != nil {
		response.Error = "invalid request"
		return
	}
	request.Amount = coin.RoundFloat(request.Amount, 5)

	// prevent users from creating too many swap requests
	if Delay.CheckUser(request.Address) {
		response.Error = "2 minutes wait time triggered"
		return
	}

	// check if pair is enabled and available
	pair := request.Pair

	if !coin.IsPairEnabled(pair) || !coin.IsPairAvailable[coin.GetPair(pair)] {
		response.Error = fmt.Sprintf("%s swap currently not possible", pair)
		return
	}

	// create swap
	err = XTCSwap(pair, &request, &response)

	if err != nil {
		response.Error = err.Error()
		log.Println(err)
	} else {
		Delay.AddUser(request.Address)
	}
	response.Request = request

	return response
}

// swap Dero to other coins
func Reverse_Swap(request coin.Swap_Request) (response coin.Swap_Response) {

	var err error

	// prevent users from creating too many swap requests
	if Delay.CheckUser(request.Address) {
		response.Error = "2 minutes wait time triggered"
		return
	}

	// check if pair is enabled and available
	pair := request.Pair

	if !coin.IsPairEnabled(pair) || !coin.IsPairAvailable[coin.GetPair(pair)] {
		response.Error = fmt.Sprintf("%s swap currently not possible", pair)
		return
	}

	response.Deposit = coin.RoundFloat(request.Amount, 5)

	// create swap
	err = DeroXTCSwap(pair, &request, &response)

	if err != nil {
		response.Error = err.Error()
		log.Println(err)
	} else {
		Delay.AddUser(request.Address)
	}
	response.Request = request

	return response
}

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()

	u.CheckOrigin = (func(r *http.Request) bool {
		return true
	})

	u.SetPongHandler(func(c *websocket.Conn, data string) {
		if err := c.SetReadDeadline(time.Now().Add(time.Minute * 2)); err != nil {
			log.Println("Pong error:", err)
		}
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		WSConnections.Delete(c)
		if v, ok := clients.Clients.LoadAndDelete(c); ok {
			log.Printf("Client Goodbye from %s\n", v.(clients.ClientInfo).Nickname)
		}
	})

	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {

		var incoming, outgoing WS_Message

		err := json.NewDecoder(bytes.NewReader(data)).Decode(&incoming)
		if err != nil {
			fmt.Println("JSON decoder:", err)
			return
		}

		if handler, ok := ws_handlers[incoming.Method]; ok {
			outgoing, err = handler(incoming, c)
			if err != nil {
				log.Println("Handler error:", err)
				return
			}
		} else {
			log.Printf("Unknown method: %s\n", incoming.Method)
			return
		}

		if outgoing.Method != "" {
			outgoing.ID = incoming.ID
			out, err := json.Marshal(&outgoing)
			if err != nil {
				log.Println("JSON:", err)
				return
			}

			err = c.WriteMessage(websocket.TextMessage, out)
			if err != nil {
				log.Println("WebSocket:", err)
			}
		}
	})

	return u
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {

	upgrader := newUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	WSConnections.Store(conn, true)
}

func WS_Broadcast(data []byte) {

	if len(data) == 0 {
		return
	}

	WSConnections.Range(func(k any, v any) bool {

		conn := k.(*websocket.Conn)

		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("WebSocket:", err)
		}

		return true
	})
}

func StartServer() {

	// cert files
	// Let's Encrypt:
	// certFile = fullchain.pem
	// keyFile = privkey.pem

	// comment the following block to disable TLS
	cert, err := tls.LoadX509KeyPair("/etc/letsencrypt/live/dero.mindmesh.de/fullchain.pem", "/etc/letsencrypt/live/dero.mindmesh.de/privkey.pem")
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", webSocketHandler)
	//mux.HandleFunc("/api", )
	set_handlers()

	srv := nbhttp.NewServer(nbhttp.Config{
		Network: "tcp",
		Handler: mux,
		// comment the following 2 lines and uncomment "Addrs" to start server without TLS
		AddrsTLS:  []string{cfg.Settings.ListenAddress},
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
		//Addrs: []string{cfg.Settings.ListenAddress},
	})

	err = srv.Start()
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	srv.Wait()
}
