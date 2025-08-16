package main

import (
	"dero-swap/clients"
	"sync"
)

type (
	Price_Provider struct {
		// TradeOgre fields
		Success bool   `json:"success"`
		Init    string `json:"initialprice"`
		Price   string `json:"price"`
		High    string `json:"high"`
		Low     string `json:"low"`
		Volume  string `json:"volume"`
		Bid     string `json:"bid"`
		Ask     string `json:"ask"`
		// CoinEx field
		Data []Price_CoinEX `json:"data"`
	}
	Price_CoinEX struct {
		Last string `json:"last"`
	}
	BTC_Fees struct {
		Fastest  uint64 `json:"fastestFee"`
		Hour     uint64 `json:"hourFee"`
		HalfHour uint64 `json:"halfHourFee"`
	}
	Pricing struct {
		Ask    float64
		Bid    float64
		Median float64
		Name   int
	}
	// Swap_Price holds the price information for a swap pair; first value is Bid, second is Ask
	Swap_Markets struct {
		BTC     []float64 `json:"btc"`
		LTC     []float64 `json:"ltc"`
		ARRR    []float64 `json:"arrr"`
		XMR     []float64 `json:"xmr"`
		BTCFees BTC_Fees  `json:"btc_fees"`
	}
	Swap_Balance struct {
		Dero     float64                 `json:"dero"`
		LTC      float64                 `json:"ltc,omitempty"`
		BTC      float64                 `json:"btc,omitempty"`
		ARRR     float64                 `json:"arrr,omitempty"`
		XMR      float64                 `json:"xmr,omitempty"`
		External []clients.Swap_External `json:"external,omitempty"`
	}
)

type MarketData struct {
	Pairs Swap_Markets
	sync.RWMutex
}

var markets = &MarketData{}
var lock sync.Mutex

type ORDER int

var PRICE_API = make(map[int][]string)
