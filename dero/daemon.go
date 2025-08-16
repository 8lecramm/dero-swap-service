package dero

import (
	"context"
	"log"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc/v3"
)

var Dero_Daemon jsonrpc.RPCClient

func IsDeroAddressRegistered(address string) bool {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if result, err := Dero_Daemon.Call(ctx, "DERO.IsRegistered", &rpc.GetEncryptedBalance_Params{Address: address, SCID: crypto.ZEROHASH}); err != nil {
		cancel()
		log.Printf("Error checking registration status: %v\n", err)
		return false
	} else {
		cancel()

		var response rpc.GetEncryptedBalance_Result

		err = result.GetObject(&response)
		if err != nil {
			log.Printf("Error checking registration status: %v\n", err)
			return false
		}

		return response.Registration > 0
	}
}

func CheckWalletBalance() (amount float64, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if result, err := Dero_Wallet.Call(ctx, "GetBalance"); err != nil {
		cancel()
		log.Printf("Error checking Dero wallet balance: %v\n", err)
		return 0, err
	} else {
		cancel()

		var response rpc.GetBalance_Result

		err = result.GetObject(&response)
		if err != nil {
			log.Printf("Error checking Dero wallet balance: %v\n", err)
			return 0, err
		}
		return float64(response.Balance) / 100000, nil
	}
}

func CheckBlockHeight() uint64 {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if result, err := Dero_Daemon.Call(ctx, "DERO.GetInfo"); err != nil {
		cancel()
		log.Printf("Error checking block height: %v\n", err)
		return 0
	} else {
		cancel()

		var response rpc.GetInfo_Result

		err = result.GetObject(&response)
		if err != nil {
			log.Printf("Error checking block height: %v\n", err)
			return 0
		}
		return uint64(response.Height)
	}
}

func DEROCheckTX(txid string) (valid bool, err error) {

	log.Println("Checking DERO transaction")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if result, err := Dero_Daemon.Call(ctx, "DERO.GetTransaction", &rpc.GetTransaction_Params{Tx_Hashes: []string{txid}}); err != nil {
		cancel()
		log.Printf("Error getting transaction: %v\n", err)
		return false, err
	} else {
		cancel()

		var response rpc.GetTransaction_Result

		err = result.GetObject(&response)
		if err != nil {
			log.Printf("Error getting transaction: %v\n", err)
			return false, err
		}

		if response.Txs[0].ValidBlock == "" || response.Txs == nil {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func CheckAddress(name string) string {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if result, err := Dero_Daemon.Call(ctx, "DERO.NameToAddress", &rpc.NameToAddress_Params{Name: name, TopoHeight: -1}); err != nil {
		cancel()
		log.Printf("Error checking name: %v\n", err)
		return ""
	} else {
		cancel()

		var response rpc.NameToAddress_Result

		err = result.GetObject(&response)
		if err != nil {
			log.Printf("Error checking name: %v\n", err)
			return ""
		}
		if response.Status == "OK" {
			return response.Address
		} else {
			return ""
		}
	}
}
