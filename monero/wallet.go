package monero

import (
	"context"
	"log"
	"time"

	"github.com/ybbus/jsonrpc/v3"
)

const atomicUnits float64 = 1000000000000

var Monero_Wallet jsonrpc.RPCClient
var RPC_Login string

func GetHeight() uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "get_height")
	cancel()

	if err != nil {
		return 0
	}

	var response RPC_XMR_Height

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error checking XMR wallet height: %v\n", err)
		return 0
	}

	return response.Height
}

func XMRGetTX(payment string, block uint64) bool {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "get_bulk_payments", RPC_XMR_BulkTX_Params{MinBlockHeight: block - 1, Payment_IDs: []string{payment}})
	cancel()

	if err != nil {
		log.Println(err)
		return false
	}

	var response RPC_XMR_GetPayments_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error checking XMR incoming payments: %v\n", err)
		return false
	}

	// todo
	for _, p := range response.Payments {
		if p.UnlockTime > 0 {
			return false
		} else {
			return true
		}
	}

	return false
}

func XMRSend(transfers []RPC_XMR_Transfer_Params) (ok bool, txid string) {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "transfer", RPC_XMR_Transfer{Destinations: transfers, Priority: 0, Ringsize: 16})
	cancel()

	if err != nil {
		return false, ""
	}

	var response RPC_XMR_Transfer_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error sending XMR transaction: %v\n%v\n", err, response)
		return false, ""
	}

	if response.TxHash != "" {
		return true, response.TxHash
	}
	return false, ""
}

func GetBalance() float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "get_balance", RPC_XMR_GetBalance_Params{AccountIndex: 0})
	cancel()

	if err != nil {
		return 0
	}

	var response RPC_XMR_GetBalance_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error checking XMR wallet balance: %v\n", err)
		return 0
	}

	return float64(response.UnlockedBalance) / atomicUnits
}

func GetAddress() string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "get_address", RPC_XMR_GetAddress_Params{AccountIndex: 0})
	cancel()

	if err != nil {
		return ""
	}

	var response RPC_XMR_GetAddress_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error checking XMR wallet address: %v\n", err)
		return ""
	}

	return response.Address
}

func MakeIntegratedAddress() string {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "make_integrated_address")
	cancel()

	if err != nil {
		return ""
	}

	var response RPC_XMR_IntegratedAddress_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error generating XMR integrated address: %v\n", err)
		return ""
	}

	return response.IntegratedAddress
}

func SplitIntegratedAddress(address string) string {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "split_integrated_address", RPC_XMR_SplitIntegratedAddress_Params{IntegratedAddress: address})
	cancel()

	if err != nil {
		return ""
	}

	var response RPC_XMR_SplitIntegratedAddress_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error splitting XMR integrated address: %v\n", err)
		return ""
	}

	return response.PaymentID
}

func ValidateAddress(address string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	result, err := Monero_Wallet.Call(ctx, "validate_address", RPC_XMR_Validate_Address_Params{Address: address})
	cancel()

	if err != nil {
		return false
	}

	var response RPC_XMR_Validate_Address_Result

	err = result.GetObject(&response)
	if err != nil {
		log.Printf("Error validating XMR address: %v\n", err)
		return false
	}

	return response.Valid
}

func AddTX(wallet string, amount float64) RPC_XMR_Transfer_Params {
	return RPC_XMR_Transfer_Params{Address: wallet, Amount: uint64(amount * atomicUnits)}
}
