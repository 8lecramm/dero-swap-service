package main

import (
	"dero-swap/cfg"
	"dero-swap/coin"
	"dero-swap/dero"
	"dero-swap/monero"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func CreateSwap(pair string, wallet string, request *coin.Swap_Request, price float64, fee uint64) int64 {

	var entry coin.Swap_Entry
	var height uint64
	var payout float64 = request.Amount

	// get current block height. Ignore transactions < height
	switch pair {
	case coin.BTC, coin.LTC, coin.ARRR:
		height = coin.XTCCheckBlockHeight(pair)
	case coin.XMR:
		height = monero.GetHeight()
	default:
		height = dero.CheckBlockHeight()
	}

	if height == 0 {
		return 0
	}

	entry.Coin = pair
	entry.Wallet = wallet
	entry.Destination = request.Address
	entry.Price = price
	entry.Amount = request.Amount
	entry.Created = time.Now().UnixMilli()
	entry.Block = height
	entry.Status = 0

	if pair == coin.DERO_BTC {
		entry.Fee = fee
	}

	// create an integrated address for all Dero -> X swaps
	if strings.HasPrefix(pair, coin.DERO) {
		if entry.Wallet = dero.MakeIntegratedAddress(entry.Created); entry.Wallet == "" {
			return 0
		}
		payout = entry.Price
	}

	json_bytes, err := json.Marshal(&entry)
	if err != nil {
		return 0
	}

	err = os.WriteFile(fmt.Sprintf("swaps/active/%d", entry.Created), json_bytes, 0644)
	if err != nil {
		return 0
	}
	coin.ActiveSwaps[entry.Created] = &entry

	log.Printf("Swap request (%d) of %.8f (%s) successfully created\n", entry.Created, payout, entry.Coin)

	push := fmt.Sprintf("%s swap created (%.8f)", entry.Coin, payout)
	PushOver(push)

	return entry.Created
}

// X to Dero Swaps
func XTCSwap(pair string, request *coin.Swap_Request, response *coin.Swap_Response) (err error) {

	// check if Dero address is registered on chain
	if !dero.IsDeroAddressRegistered(request.Address) {
		return fmt.Errorf("dero address is not registered")
	}

	// check balance and include locked swap balance
	// TODO: create a new function to check balance
	balance, err := dero.CheckWalletBalance()
	if err != nil {
		return fmt.Errorf("couldn't check swap balance")
	}
	if coin.LockedBalance.GetLockedBalance(pair)+request.Amount+dero.TxFee > balance {
		return fmt.Errorf("insufficient swap balance")
	}

	coin.LockedBalance.AddLockedBalance(pair, request.Amount)

	// create/get a deposit address
	switch pair {
	case coin.BTC:
		response.Wallet = coin.BTC_address
	case coin.LTC:
		response.Wallet = coin.LTC_address
	case coin.XMR:
		response.Wallet = monero.MakeIntegratedAddress()
	case coin.ARRR:
		response.Wallet = coin.ARRR_address
	}
	if response.Wallet == "" {
		return fmt.Errorf("no swap deposit address available")
	}

	var coin_value float64
	var atomicUnits uint = 8

	switch pair {
	case coin.BTC:
		coin_value = markets.Pairs.BTC[ASK]
	case coin.LTC:
		coin_value = markets.Pairs.LTC[ASK]
	case coin.ARRR:
		coin_value = markets.Pairs.ARRR[ASK]
	case coin.XMR:
		coin_value = markets.Pairs.XMR[ASK]
		atomicUnits = 12
	}

	var deposit_value float64
	if cfg.Settings.Mode == cfg.SERVER {
		deposit_value = coin_value * request.Amount
	} else {
		deposit_value = request.Price
	}

	var loops int = 5
	var isAvailable bool

	// if there is a request with the same deposit amount, run in a loop and lower deposit value by 1 Sat
	for i := 0; i < loops; i++ {
		if coin.IsAmountFree(pair, deposit_value) {
			isAvailable = true
			break
		}
		deposit_value -= SATOSHI
	}
	if !isAvailable || deposit_value == 0 {
		return fmt.Errorf("Pre-Check: Couldn't create swap")
	}

	deposit_value = coin.RoundFloat(deposit_value, atomicUnits)

	response.ID = CreateSwap(pair, response.Wallet, request, deposit_value, 0)
	if response.ID == 0 {
		return fmt.Errorf("couldn't create swap")
	}
	response.Deposit = deposit_value

	return nil
}

// Dero to X swaps
func DeroXTCSwap(pair string, request *coin.Swap_Request, response *coin.Swap_Response) (err error) {

	var balance float64
	var atomicUnits uint = 8

	// validate destination wallet and check for sufficient swap balance
	if pair != coin.DERO_XMR {
		if !coin.XTCValidateAddress(pair, request.Address) {
			return fmt.Errorf("%s address is not valid", pair[5:])
		}
		balance = coin.XTCGetBalance(pair)
	} else {
		if !monero.ValidateAddress(request.Address) {
			return fmt.Errorf("XMR address is not valid")
		}
		balance = monero.GetBalance()
		atomicUnits = 12
	}

	var coin_value float64
	var fees float64
	var btc_fees BTC_Fees

	// determine fees and current price
	switch pair {
	case coin.DERO_LTC:
		coin_value = markets.Pairs.LTC[BID]
		fees = cfg.SwapFees.Withdrawal.LTC
	case coin.DERO_ARRR:
		coin_value = markets.Pairs.ARRR[BID]
		fees = cfg.SwapFees.Withdrawal.ARRR
	case coin.DERO_BTC:
		coin_value = markets.Pairs.BTC[BID]
		//fees = cfg.SwapFees.Withdrawal.DeroBTC
		btc_fees, _ = GetBTCFees()
		fees = float64((btc_fees.HalfHour*141)+(500-((btc_fees.HalfHour*141)%500))) / 100000000

	case coin.DERO_XMR:
		coin_value = markets.Pairs.XMR[BID]
		fees = cfg.SwapFees.Withdrawal.XMR
	}

	var payout_value float64
	if cfg.Settings.Mode == cfg.SERVER {
		payout_value = coin_value * request.Amount
	} else {
		payout_value = request.Price
	}

	if payout_value-fees <= 0 {
		return fmt.Errorf("fees > payout value")
	}
	if payout_value == 0 || fees == 0 {
		return fmt.Errorf("couldn't create swap")
	}

	// check for reserved balance
	if coin.LockedBalance.GetLockedBalance(pair)+payout_value+fees > balance {
		return fmt.Errorf("insufficient swap balance")
	}

	coin.LockedBalance.AddLockedBalance(pair, payout_value)

	payout_value -= fees
	payout_value = coin.RoundFloat(payout_value, atomicUnits)
	response.Swap = payout_value

	response.ID = CreateSwap(pair, "", request, payout_value, btc_fees.HalfHour)
	if response.ID == 0 {
		return fmt.Errorf("couldn't create swap")
	}
	response.Wallet = dero.MakeIntegratedAddress(response.ID)

	return nil
}

func GetPayoutValue(pair string, amount float64) float64 {

	var coin_value float64
	var fees float64
	var btc_fees BTC_Fees

	// determine fees and current price
	switch pair {
	case coin.DERO_LTC:
		coin_value = markets.Pairs.LTC[BID]
		fees = cfg.SwapFees.Withdrawal.LTC
	case coin.DERO_ARRR:
		coin_value = markets.Pairs.ARRR[BID]
		fees = cfg.SwapFees.Withdrawal.ARRR
	case coin.DERO_BTC:
		coin_value = markets.Pairs.BTC[BID]
		btc_fees, _ = GetBTCFees()
		fees = float64((btc_fees.HalfHour*141)+(500-((btc_fees.HalfHour*141)%500))) / 100000000
	case coin.DERO_XMR:
		coin_value = markets.Pairs.XMR[BID]
		fees = cfg.SwapFees.Withdrawal.XMR
	default:
		coin_value = 1
	}

	if fees > (coin_value * amount) {
		return 0
	}
	return (coin_value * amount) - fees
}
