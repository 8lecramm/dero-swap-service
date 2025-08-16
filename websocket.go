package main

import (
	"dero-swap/cfg"
	"dero-swap/clients"
	"dero-swap/coin"
	"dero-swap/dero"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lesismal/nbio/nbhttp/websocket"
)

func set_handlers() {
	ws_handlers["market"] = MarketHandler
	ws_handlers["balance"] = BalanceHandler
	ws_handlers["swap"] = SwapHandler
	ws_handlers["client"] = NewClientHandler
	ws_handlers["client_ok"] = ClientResponseHandler
}

func BalanceHandler(msg WS_Message, conn *websocket.Conn) (result WS_Message, err error) {
	result.Method = "balance"
	result.Result = UpdatePool()

	return result, nil
}

func MarketHandler(msg WS_Message, conn *websocket.Conn) (result WS_Message, err error) {
	markets.RLock()
	defer markets.RUnlock()

	result.Method = "market"
	result.Result = markets.Pairs

	return result, nil
}

func SwapHandler(msg WS_Message, conn *websocket.Conn) (result WS_Message, err error) {
	var request coin.Swap_Request

	var user string
	var forward bool

	p := msg.Params.(map[string]any)

	if pair, ok := p["pair"].(string); ok {
		request.Pair = pair
	}
	if amount, ok := p["amount"].(float64); ok {
		request.Amount = amount
	}
	// TODO: Update webpage to use "address" instead of "dero_address"
	if address, ok := p["dero_address"].(string); ok {
		request.Address = address
	}
	if client, ok := p["client"].(string); ok {
		user = client
	}
	if cfg.Settings.Mode == cfg.CLIENT {
		if d, ok := p["price"].(float64); ok {
			request.Price = d
		}
	}

	if cfg.Settings.Mode == cfg.SERVER {
		// forward some requests to a partner
		if (time.Now().UnixMilli() % 2) == 0 {
			forward = true
		}

		// check if swap can be handled by a partner if the requested amount is too high for the main pool
		// TODO: simplify, eventually create a new function
		amount := GetPayoutValue(request.Pair, request.Amount)
		if strings.HasSuffix(request.Pair, coin.DERO) {
			if balance, err := dero.CheckWalletBalance(); err == nil {
				if coin.LockedBalance.GetLockedBalance(request.Pair)+amount+dero.TxFee > balance {
					ext := clients.GetExternalBalances()
					for _, e := range ext {
						if e.Dero >= amount {
							user = e.Nickname
							forward = true
							break
						}
					}
				}
			}
		}

		// TODO: refactorize; client must process concurrent swaps
		if (user != "" || forward) && amount > 0 {
			if user == "" {
				if user = clients.GetActiveClient(request.Pair, amount); user != "" {
					log.Printf("%s has been chosen to handle the swap\n", user)
				}
			}
			if user != "" {
				if ok, c := clients.PrepareExternalSwap(user, request.Pair, amount); ok {
					clients.ActiveClients.ChangeClientState(clients.LOCK, c)
					clients.ActiveClients.AddOrigin(c, conn)

					request.Price = amount
					data, _ := json.Marshal(WS_Message{
						ID:     msg.ID,
						Method: "swap",
						Params: request,
					})

					err = c.WriteMessage(websocket.TextMessage, data)
					if err != nil {
						log.Println("Error sending websocket message:", err)
					}

					return
				}
			}
			log.Println("No suitable partner found or swap cannot be handled by partner")
		}
	}

	isReverse := false
	if strings.HasPrefix(request.Pair, "dero") {
		isReverse = true
	}

	var response coin.Swap_Response
	if !isReverse {
		response = Dero_Swap(request)
	} else {
		response = Reverse_Swap(request)
	}

	result.Method = "swap"
	result.Result = response

	return result, nil
}

func NewClientHandler(msg WS_Message, conn *websocket.Conn) (result WS_Message, err error) {
	var client clients.ClientInfo

	p := msg.Params.(map[string]any)
	client.Nickname = p["nickname"].(string)

	if c, ok := p["pair_info"]; ok && c != nil {
		q := c.([]any)

		for i := range q {
			r := q[i].(map[string]any)
			client.PairInfo = append(client.PairInfo, clients.PairInfo{
				Balance: r["balance"].(float64),
				Pair:    r["pair"].(string)})
		}
	}
	if _, ok := clients.Clients.Load(conn); !ok {
		log.Printf("Client Hello from %s\n", client.Nickname)
	}
	clients.Clients.Store(conn, client)

	return result, nil
}

func ClientResponseHandler(msg WS_Message, conn *websocket.Conn) (result WS_Message, err error) {
	if clients.ActiveClients.CheckClientState(conn) {

		clients.ActiveClients.ChangeClientState(clients.UNLOCK, conn)
		client := clients.ActiveClients.GetOrigin(conn)

		result.Method = "swap"
		result.Result = msg.Result

		if d, ok := clients.Clients.Load(conn); ok {
			msg := fmt.Sprintf("Swap request accepted by %s", d.(clients.ClientInfo).Nickname)
			log.Println(msg)
			//PushOver(msg)
		}

		data, _ := json.Marshal(result)
		err = client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("Error sending websocket message:", err)
		}
		result.Method = ""

		return result, nil
	}

	return result, fmt.Errorf("client not locked or not found")
}
