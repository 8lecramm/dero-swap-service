package clients

import (
	"sync"

	"github.com/lesismal/nbio/nbhttp/websocket"
)

const (
	LOCK = iota
	UNLOCK
)

type (
	ClientInfo struct {
		PairInfo []PairInfo `json:"pair_info"`
		Nickname string     `json:"nickname"`
	}
	PairInfo struct {
		Pair    string  `json:"pair"`
		Balance float64 `json:"balance"`
	}
)
type SwapState struct {
	Client map[*websocket.Conn]bool
	Result map[*websocket.Conn]*websocket.Conn
	sync.RWMutex
}

var Clients sync.Map
var ActiveClients = SwapState{
	Client: make(map[*websocket.Conn]bool),
	Result: make(map[*websocket.Conn]*websocket.Conn),
}
var ActiveClient string
