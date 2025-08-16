package clients

import (
	"dero-swap/cfg"
	"dero-swap/coin"
	"log"

	"github.com/lesismal/nbio/nbhttp/websocket"
)

type Swap_External struct {
	Nickname string  `json:"nickname"`
	Dero     float64 `json:"dero"`
	XMR      float64 `json:"xmr,omitempty"`
	LTC      float64 `json:"ltc,omitempty"`
	BTC      float64 `json:"btc,omitempty"`
	ARRR     float64 `json:"arrr,omitempty"`
}

func IsExternalSwapAvailable(user string, pair string, amount float64) (ok bool, client *websocket.Conn) {

	pair = coin.GetBalancePair(pair)

	Clients.Range(func(key any, value any) bool {
		c := value.(ClientInfo)
		for _, p := range c.PairInfo {
			if p.Pair == pair && p.Balance > amount && user == c.Nickname {
				ok = true
				client = key.(*websocket.Conn)
				return false
			}
		}
		return true
	})
	return
}

func PrepareExternalSwap(user string, pair string, amount float64) (bool, *websocket.Conn) {

	ok, conn := IsExternalSwapAvailable(user, pair, amount)
	if !ok {
		log.Printf("No 3rd party swap for user %s available\n", user)
		return false, nil
	}

	return true, conn
}

func (c *SwapState) ChangeClientState(mode uint, conn *websocket.Conn) {

	c.Lock()
	defer c.Unlock()

	if mode == LOCK {
		c.Client[conn] = true
	} else {
		c.Client[conn] = false
	}
}

func (c *SwapState) CheckClientState(conn *websocket.Conn) bool {

	c.RLock()
	defer c.RUnlock()

	return c.Client[conn]
}

func (c *SwapState) AddOrigin(conn *websocket.Conn, target *websocket.Conn) {

	c.Lock()
	defer c.Unlock()

	c.Result[conn] = target
}

func (c *SwapState) GetOrigin(conn *websocket.Conn) *websocket.Conn {

	c.RLock()
	defer c.RUnlock()

	return c.Result[conn]
}

func GetExternalBalances() (list []Swap_External) {

	var entry Swap_External

	Clients.Range(func(key any, value any) bool {
		c := value.(ClientInfo)
		entry.Nickname = c.Nickname
		for _, p := range c.PairInfo {
			switch p.Pair {
			case cfg.LTC:
				entry.LTC = p.Balance
			case cfg.BTC:
				entry.BTC = p.Balance
			case cfg.XMR:
				entry.XMR = p.Balance
			case cfg.ARRR:
				entry.ARRR = p.Balance
			case coin.DERO:
				entry.Dero = p.Balance

			}
		}
		list = append(list, entry)
		return true
	})

	return
}

func GetActiveClient(pair string, amount float64) (user string) {

	if amount == 0 {
		return
	}
	pair = coin.GetBalancePair(pair)

	Clients.Range(func(key any, value any) bool {
		c := value.(ClientInfo)
		for _, p := range c.PairInfo {
			if p.Pair == pair && p.Balance > amount {
				user = c.Nickname
				return false
			}
		}
		return true
	})

	return
}
