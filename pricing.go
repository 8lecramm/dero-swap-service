package main

import (
	"bytes"
	"dero-swap/cfg"
	"dero-swap/clients"
	"dero-swap/coin"
	"dero-swap/dero"
	"dero-swap/monero"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func Init_Pricing() {
	PRICE_API[DERO_USDT] = []string{
		"https://tradeogre.com/api/v1/ticker/DERO-USDT",
		"https://api.coinex.com/v2/spot/ticker?market=DEROUSDT",
	}
	PRICE_API[DERO_BTC] = []string{
		"https://tradeogre.com/api/v1/ticker/DERO-BTC",
		"https://api.coinex.com/v2/spot/ticker?market=DEROBTC",
	}
	PRICE_API[LTC_USDT] = []string{
		"https://tradeogre.com/api/v1/ticker/LTC-USDT",
		"https://api.coinex.com/v2/spot/ticker?market=LTCUSDT",
	}
	PRICE_API[XMR_USDT] = []string{
		"https://tradeogre.com/api/v1/ticker/XMR-USDT",
		"https://api.coinex.com/v2/spot/ticker?market=XMRUSDT",
	}
	PRICE_API[ARRR_USDT] = []string{
		"https://tradeogre.com/api/v1/ticker/ARRR-USDT",
		"https://api.coinex.com/v2/spot/ticker?market=ARRRUSDT",
	}

	// not necessary, but the current swap web page needs to be updated first
	zeroPrice := []float64{0, 0}
	markets.Pairs.BTC = zeroPrice
	markets.Pairs.LTC = zeroPrice
	markets.Pairs.ARRR = zeroPrice
	markets.Pairs.XMR = zeroPrice
}

func GetMarket(url string) (result float64) {

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}

	var market Price_Provider

	if err := json.Unmarshal(body, &market); err != nil {
		return 0
	}

	if strings.Contains(url, "tradeogre.com") {
		result, err = strconv.ParseFloat(market.Price, 64)
		if err != nil {
			log.Println("Market: Cannot convert Price:", err)
		}
	} else {
		result, err = strconv.ParseFloat(market.Data[0].Last, 64)
		if err != nil {
			log.Println("Market: Cannot convert Ask:", err)
		}
	}
	if err != nil {
		result = 0
		log.Println("Market: Cannot convert Price:", err)
	}

	return
}

// Get Pair values
func GetPrice(pair string) float64 {

	var price []float64
	var providers []string

	switch pair {
	case cfg.BTC:
		providers = PRICE_API[DERO_BTC]
	case cfg.LTC:
		providers = PRICE_API[LTC_USDT]
	case cfg.XMR:
		providers = PRICE_API[XMR_USDT]
	case cfg.ARRR:
		providers = PRICE_API[ARRR_USDT]
	case coin.DERO:
		providers = PRICE_API[DERO_USDT]
	}

	for _, u := range providers {
		price = append(price, GetMarket(u))
	}
	if len(price) == 0 {
		log.Println("GetPrice: no price available for", pair)
		return 0
	}

	return slices.Max(price)
}

func UpdateMarkets() {

	var last float64
	var bid, ask float64
	var atomicUnits uint

	base_usd := GetPrice(coin.DERO)
	if base_usd == 0 {
		log.Println("UpdateMarkets: no base price available")
		return
	}
	usd_bid := base_usd - (base_usd * cfg.SwapFees.Fees / 100)
	usd_ask := base_usd + (base_usd * cfg.SwapFees.Fees / 100)

	markets.Lock()
	defer markets.Unlock()

	for p := range coin.Pairs {

		last = GetPrice(p)

		if last == 0 {
			log.Printf("UpdateMarkets: %s: no price available\n", p)
			log.Println(p, "temporarily disabled")
			coin.IsPairAvailable[p] = false

			continue
		}
		if p == cfg.XMR {
			atomicUnits = 12
		} else {
			atomicUnits = 8
		}
		if p != cfg.BTC {
			bid = coin.RoundFloat(usd_bid/last, atomicUnits)
			ask = coin.RoundFloat(usd_ask/last, atomicUnits)
		} else {
			bid = coin.RoundFloat(last-(last*cfg.SwapFees.Fees/100), atomicUnits)
			ask = coin.RoundFloat(last+(last*cfg.SwapFees.Fees/100), atomicUnits)
		}

		if err := cfg.SetFieldByJSONTag(&markets.Pairs, p, []float64{bid, ask}); err != nil {
			log.Printf("UpdateMarkets: %s: %v\n", p, err)
			continue
		}
		coin.IsPairAvailable[p] = true
	}

	var ok bool
	if markets.Pairs.BTCFees, ok = GetBTCFees(); !ok {
		log.Println("UpdateMarkets: cannot get BTC fees")
		return
	}

	var data bytes.Buffer
	var message WS_Message

	message.Method = "market"
	message.Result = markets.Pairs
	encoder := json.NewEncoder(&data)

	err := encoder.Encode(message)
	if err != nil {
		log.Println("Market:", err)
		return
	}
	WS_Broadcast(data.Bytes())
	data.Reset()

	message.Method = "balance"
	message.Result = UpdatePool()
	err = encoder.Encode(message)
	if err != nil {
		log.Println("Balance:", err)
		return
	}
	WS_Broadcast(data.Bytes())
}

// TODO: reduce function calls
func UpdatePool() Swap_Balance {

	lock.Lock()
	defer lock.Unlock()

	var balance Swap_Balance
	var info clients.ClientInfo

	balance.Dero = dero.GetBalance()

	for p := range coin.Pairs {
		value := float64(0)
		switch p {
		case cfg.XMR:
			value = monero.GetBalance()
			if err := cfg.SetFieldByJSONTag(&balance, p, value); err != nil {
				log.Println("UpdatePool:", err)
				continue
			}
		default:
			value = coin.XTCGetBalance(p)
			if err := cfg.SetFieldByJSONTag(&balance, p, value); err != nil {
				log.Println("UpdatePool:", err)
				continue
			}
		}
		if cfg.Settings.Mode == cfg.CLIENT {
			info.PairInfo = append(info.PairInfo, clients.PairInfo{
				Pair:    p,
				Balance: value,
			})
		}
	}

	if cfg.Settings.Mode == cfg.SERVER {
		balance.External = clients.GetExternalBalances()

		// swappers might overlook 3rd party Dero balance, so we display the highest possible swap amount
		if len(balance.External) > 0 {
			var external_dero_balance []float64
			for _, b := range balance.External {
				external_dero_balance = append(external_dero_balance, b.Dero)
			}
			max_possible := slices.Max(external_dero_balance)
			if max_possible > balance.Dero {
				balance.Dero = max_possible
			}
		}
	} else {
		info.PairInfo = append(info.PairInfo, clients.PairInfo{
			Pair:    coin.DERO,
			Balance: balance.Dero,
		})
		info.Nickname = cfg.Settings.Nickname

		if Connection != nil {
			if err := Connection.WriteJSON(WS_Message{
				Method: "client",
				Params: info,
			}); err != nil {
				log.Println("UpdatePool: cannot send client info:", err)
			}
		}
	}

	return balance
}

func GetBTCFees() (fees BTC_Fees, ok bool) {

	resp, err := http.Get(BTC_MEMPOOL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err := json.Unmarshal(body, &fees); err != nil {
		return
	}
	if fees.Fastest == 0 || fees.Hour == 0 || fees.HalfHour == 0 {
		return
	}

	return fees, true
}
