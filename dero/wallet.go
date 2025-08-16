package dero

import (
	"context"
	"dero-swap/coin"
	"log"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc/v3"
)

const atomicUnits float64 = 100000
const TxFee float64 = 0.0008

var RPC_Login string

var Dero_Wallet jsonrpc.RPCClient

func AddTX(wallet string, amount float64) rpc.Transfer {
	return rpc.Transfer{SCID: crypto.ZEROHASH, Destination: wallet, Amount: uint64(amount * atomicUnits)}
}

func GetHeight() uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Dero_Wallet.Call(ctx, "getheight")
	cancel()

	if err != nil {
		return 0
	}

	var response rpc.GetHeight_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error checking DERO wallet height: %v\n", err)
		return 0
	}

	return response.Height
}

func GetBalance() float64 {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Dero_Wallet.Call(ctx, "getbalance")
	cancel()

	if err != nil {
		return 0
	}

	var response rpc.GetBalance_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error checking DERO wallet balance: %v\n", err)
		return 0
	}

	balance_fl := float64(response.Balance) / atomicUnits
	balance_fl -= coin.LockedBalance.GetLockedBalance(coin.BTC)

	return coin.RoundFloat(balance_fl, 5)
}

func GetAddress() string {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Dero_Wallet.Call(ctx, "getaddress")
	cancel()

	if err != nil {
		return ""
	}

	var response rpc.GetAddress_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error checking DERO wallet address: %v\n", err)
		return ""
	}

	return response.Address
}

func Payout(tx []rpc.Transfer) {

	var tries uint
	var reserve float64

	for _, e := range tx {
		reserve += float64(e.Amount) / atomicUnits
	}

	for {
		time.Sleep(time.Second * 18)

		tries++
		if tries > 1 {
			log.Println("Too many failures, manual check necessary!")
			break
		}

		log.Printf("Sending DERO transaction try #%d\n", tries)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		if result, err := Dero_Wallet.Call(ctx, "Transfer", &rpc.Transfer_Params{Ringsize: 16, Fees: 80, Transfers: tx}); err != nil {
			cancel()
			log.Printf("Error sending DERO transaction: %v\n", err)
			continue
		} else {
			cancel()

			var response rpc.Transfer_Result

			err = result.GetObject(&response)
			if err != nil {
				log.Printf("Error sending DERO transaction: %v\n", err)
				continue
			}
			if response.TXID == "" {
				continue
			}

			var init_block, current_block uint64
			var retry bool

			for init_block == 0 {
				init_block = CheckBlockHeight()
				time.Sleep(time.Second * 3)
			}
			current_block = init_block

			retry = true
			for current_block-init_block <= 12 {
				ok, err := DEROCheckTX(response.TXID)
				if err != nil {
					time.Sleep(time.Second * 3)
					continue
				}
				if ok {
					log.Printf("DERO transaction (TXID %s) successfully sent\n", response.TXID)
					coin.LockedBalance.RemoveLockedBalance(coin.BTC, reserve)
					retry = false
					break
				}
				current_block = 0
				for current_block == 0 {
					current_block = CheckBlockHeight()
					time.Sleep(time.Second * 3)
				}
				time.Sleep(time.Second * 15)
			}
			if retry {
				continue
			}
			break
		}
	}
}

func CheckIncomingTransfers(dstPort uint64, block uint64) bool {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Dero_Wallet.Call(ctx, "GetTransfers", &rpc.Get_Transfers_Params{In: true, Min_Height: block})
	cancel()
	if err != nil {
		log.Printf("Error sending DERO request: %v\n", err)
		return false
	}

	var response rpc.Get_Transfers_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error retrieving DERO response: %v\n", err)
		return false
	}

	for _, e := range response.Entries {
		if e.DestinationPort == dstPort {
			return true
		}
	}

	return false
}

func MakeIntegratedAddress(sessionID int64) string {

	var payload = rpc.Arguments{rpc.Argument{Name: "D", DataType: "U", Value: uint64(sessionID)}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := Dero_Wallet.Call(ctx, "MakeIntegratedAddress", &rpc.Make_Integrated_Address_Params{Payload_RPC: payload})
	cancel()
	if err != nil {
		log.Printf("Error sending DERO request: %v\n", err)
		return ""
	}

	var response rpc.Make_Integrated_Address_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error retrieving DERO response: %v\n", err)
		return ""
	}

	return response.Integrated_Address
}
