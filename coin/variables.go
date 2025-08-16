package coin

import (
	"net/http"
	"sync"
)

type (
	Swap_Request struct {
		Pair    string  `json:"pair"`
		Amount  float64 `json:"amount"`
		Address string  `json:"address"`
		Price   float64 `json:"price,omitempty"`
		Client  string  `json:"client,omitempty"`
	}
	Swap_Response struct {
		ID      int64        `json:"id"`
		Wallet  string       `json:"wallet,omitempty"`
		Deposit float64      `json:"deposit,omitempty"`
		Swap    float64      `json:"swap,omitempty"`
		Error   string       `json:"error,omitempty"`
		Request Swap_Request `json:"request"`
	}
	Swap_Tracking struct {
		ID    int64  `json:"id"`
		State uint64 `json:"state"`
	}
	Swap_Entry struct {
		Coin        string  `json:"coin"`
		Wallet      string  `json:"wallet"`
		Destination string  `json:"destination"`
		Amount      float64 `json:"amount"`
		Price       float64 `json:"price"`
		Fee         uint64  `json:"fee,omitempty"`
		Created     int64   `json:"created"`
		Block       uint64  `json:"block"`
		Balance     float64 `json:"balance,omitempty"`
		Status      uint64  `json:"status"`
		Txid        string  `json:"txid,omitempty"`
	}
	Swap struct {
		Dero float64
		LTC  float64
		BTC  float64
		ARRR float64
		XMR  float64
		sync.RWMutex
	}
)

var LockedBalance Swap
var ActiveSwaps = make(map[int64]*Swap_Entry)

var XTC_URL = make(map[string]string)
var XTC_Daemon = &http.Client{}
var BTC_Login, LTC_Login string

var BTC_address, LTC_address, ARRR_address, XMR_address string
var BTC_Dir, LTC_Dir, ARRR_Dir string

var Pairs = make(map[string]bool)
var IsPairAvailable = make(map[string]bool)
