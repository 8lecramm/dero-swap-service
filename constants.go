package main

// BTC mempool fees
const BTC_MEMPOOL = "https://mempool.space/api/v1/fees/recommended"

const (
	BID = iota
	ASK
)

const (
	TO = iota
	COINEX
)

const SATOSHI float64 = 1e-08

const (
	DERO_USDT = iota
	DERO_BTC
	LTC_USDT
	XMR_USDT
	ARRR_USDT
)
