package coin

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"strings"
)

func GetPair(pair string) string {
	if strings.HasSuffix(pair, DERO) {
		pair, _ = strings.CutSuffix(pair, "-dero")
	} else {
		pair, _ = strings.CutPrefix(pair, "dero-")
	}

	return pair
}

func GetBalancePair(pair string) string {
	if strings.HasSuffix(pair, DERO) {
		pair = DERO
	} else {
		pair, _ = strings.CutPrefix(pair, "dero-")
	}

	return pair
}

func IsPairEnabled(coin string) bool {
	coin = GetPair(coin)
	for i := range Pairs {
		if coin == i {
			return true
		}
	}
	return false
}

func IsAmountFree(coin string, amount float64) bool {

	for _, e := range ActiveSwaps {
		if e.Price == amount && e.Coin == coin {
			return false
		}
	}

	return true
}

func (r *Swap) GetLockedBalance(coin string) float64 {
	r.RLock()
	defer r.RUnlock()

	switch coin {
	case BTC, LTC, ARRR, XMR:
		return r.Dero
	case DERO_LTC:
		return r.LTC
	case DERO_BTC:
		return r.BTC
	case DERO_ARRR:
		return r.ARRR
	case DERO_XMR:
		return r.XMR
	default:
		return 0
	}
}

func (r *Swap) AddLockedBalance(coin string, amount float64) {
	r.Lock()
	defer r.Unlock()

	switch coin {
	case BTC, LTC, ARRR, XMR:
		r.Dero += amount
	case DERO_LTC:
		r.LTC += amount
	case DERO_BTC:
		r.BTC += amount
	case DERO_ARRR:
		r.ARRR += amount
	case DERO_XMR:
		r.XMR += amount
	}
}

func (r *Swap) RemoveLockedBalance(coin string, amount float64) {
	r.Lock()
	defer r.Unlock()

	switch coin {
	case BTC, LTC, ARRR, XMR:
		r.Dero -= amount
	case DERO_LTC:
		r.LTC -= amount
	case DERO_BTC:
		r.BTC -= amount
	case DERO_ARRR:
		r.ARRR -= amount
	case DERO_XMR:
		r.XMR -= amount
	}
}

func (r *Swap) LoadActiveSwaps() {

	dir_entries, err := os.ReadDir("swaps/active")
	if err != nil {
		ErrorCheckingOpenSwaps()
	}

	var swap_e Swap_Entry
	var count int

	for _, e := range dir_entries {
		file_data, err := os.ReadFile("swaps/active/" + e.Name())
		if err != nil {
			ErrorCheckingOpenSwaps()
		}
		err = json.Unmarshal(file_data, &swap_e)
		if err != nil {
			ErrorCheckingOpenSwaps()
		}
		if !strings.HasPrefix(swap_e.Coin, DERO) {
			r.AddLockedBalance(swap_e.Coin, swap_e.Amount)
		} else {
			r.AddLockedBalance(swap_e.Coin, swap_e.Price)
		}
		ActiveSwaps[swap_e.Created] = &Swap_Entry{
			Coin:        swap_e.Coin,
			Wallet:      swap_e.Wallet,
			Destination: swap_e.Destination,
			Amount:      swap_e.Amount,
			Price:       swap_e.Price,
			Fee:         swap_e.Fee,
			Created:     swap_e.Created,
			Block:       swap_e.Block,
			Balance:     swap_e.Balance,
			Status:      swap_e.Status,
			Txid:        swap_e.Txid,
		}
		count++
	}
	log.Println("Loaded", count, "active swaps")
}

func ErrorCheckingOpenSwaps() {
	log.Println("Can't check reserved amounts")
	os.Exit(1)
}

// round value to X decimal places
func RoundFloat(value float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(value*ratio) / ratio
}
