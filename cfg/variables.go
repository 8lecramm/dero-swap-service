package cfg

import "net/url"

type Config struct {
	ListenAddress string   `json:"listen"`
	ServerAddress string   `json:"server"`
	Nickname      string   `json:"nickname"`
	BTC_Daemon    string   `json:"BTC_Daemon"`
	BTC_Dir       string   `json:"BTC_Dir"`
	BTC_Login     string   `json:"BTC_Login"`
	LTC_Daemon    string   `json:"LTC_Daemon"`
	LTC_Dir       string   `json:"LTC_Dir"`
	LTC_Login     string   `json:"LTC_Login"`
	ARRR_Daemon   string   `json:"ARRR_Daemon"`
	ARRR_Dir      string   `json:"ARRR_Dir"`
	Dero_Daemon   string   `json:"Dero_Daemon"`
	Dero_Wallet   string   `json:"dero_wallet"`
	Dero_Login    string   `json:"dero_login"`
	Monero_Daemon string   `json:"monero_daemon"`
	Monero_Wallet string   `json:"Monero_Wallet"`
	Monero_Login  string   `json:"monero_login"`
	Pairs         []string `json:"pairs"`
	// dynamically updated
	Mode      int
	ServerURL url.URL
}

type (
	Fees struct {
		Fees       float64         `json:"fees"`
		Withdrawal Withdrawal_Fees `json:"withdrawal"`
	}
	Swap_Fees struct {
		Bid float64 `json:"bid"`
		Ask float64 `json:"ask"`
	}
	Withdrawal_Fees struct {
		LTC  float64 `json:"ltc"`
		BTC  float64 `json:"btc"`
		ARRR float64 `json:"arrr"`
		XMR  float64 `json:"xmr"`
	}
)

const (
	CLIENT = iota
	SERVER
)

const (
	XMR  = "xmr"
	BTC  = "btc"
	LTC  = "ltc"
	ARRR = "arrr"
)

var Settings Config
var SwapFees Fees

var SupportedCoins = []string{
	XMR,
	BTC,
	LTC,
	ARRR,
}
