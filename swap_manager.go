package main

import (
	"dero-swap/coin"
	"dero-swap/dero"
	"dero-swap/monero"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deroproject/derohe/rpc"
)

const (
	SWAP_CREATED = iota
	SWAP_CONFIRMED
	SWAP_DONE
	eXPIRED
)

func Swap_Controller() {

	var file_data []byte
	var expired, fails, sent uint
	var active int

	var txs []rpc.Transfer
	var xmr_txs []monero.RPC_XMR_Transfer_Params

	var err error

	for {

		time.Sleep(time.Minute)

		expired, fails = 0, 0
		sent, active = 0, 0
		txs, xmr_txs = nil, nil

		for t, e := range coin.ActiveSwaps {
			err = nil

			active = len(coin.ActiveSwaps)
			file_data, err = json.Marshal(&e)
			if err != nil {
				log.Printf("Error marshalling swap entry %d: %v\n", t, err)
				fails++
				continue
			}

			// if there was no deposit, mark the request as expired
			if e.Status == SWAP_CREATED && time.Since(time.UnixMilli(t)) > time.Hour {
				os.WriteFile(fmt.Sprintf("swaps/expired/%d", t), file_data, 0644)
				os.Remove(fmt.Sprintf("swaps/active/%d", t))
				delete(coin.ActiveSwaps, t)
				switch e.Coin {
				case coin.LTC, coin.BTC, coin.ARRR, coin.XMR:
					coin.LockedBalance.RemoveLockedBalance(e.Coin, e.Amount)
				default:
					coin.LockedBalance.RemoveLockedBalance(e.Coin, e.Price)
				}

				expired++
				continue
			}

			var found_deposit, visible bool

			// check for deposits
			switch e.Coin {
			case coin.BTC, coin.LTC, coin.ARRR:
				found_deposit, visible, _, err = coin.XTCListReceivedByAddress(e.Coin, e.Wallet, e.Price, e.Block, false)
			case coin.XMR:
				if payment_id := monero.SplitIntegratedAddress(e.Wallet); payment_id != "" {
					found_deposit = monero.XMRGetTX(payment_id, e.Block)
					visible = found_deposit
				} else {
					log.Println("Can't split integrated XMR address")
				}
			default:
				found_deposit = dero.CheckIncomingTransfers(uint64(t), e.Block)
				visible = found_deposit
			}

			if err != nil {
				log.Printf("Error checking incoming %s transactions\n", e.Coin)
				fails++
				continue
			}

			// mark request as done
			if e.Status == SWAP_DONE {
				err = os.WriteFile(fmt.Sprintf("swaps/done/%d", t), file_data, 0644)
				if err != nil {
					log.Printf("Can't mark swap as done, swap %d, err %v\n", t, err)
				} else {
					delete(coin.ActiveSwaps, t)
				}
			}

			// start payout if there are at least 2 confirmations
			// requests won't be marked as expired, if there is already 1 confirmation
			if visible {
				log.Printf("Found TX for ID %d (%s) on chain\n", t, e.Coin)
				if found_deposit && e.Status <= SWAP_CONFIRMED {
					// create transaction
					log.Printf("Found deposit for ID %d (%s): %.8f coins; adding to payout TX\n", t, e.Coin, e.Amount)

					switch e.Coin {
					// TODO: support "sendtomany" call
					case coin.DERO_LTC, coin.DERO_BTC:
						log.Println("Starting LTC/BTC payout")
						_, txid := coin.XTCSend(e.Coin, e.Destination, e.Price, e.Fee)
						log.Printf("LTC/BTC TXID: %s\n", txid)
						coin.LockedBalance.RemoveLockedBalance(e.Coin, e.Price)

						push := fmt.Sprintf("%s swap done", e.Coin)
						PushOver(push)
					case coin.DERO_ARRR:
						log.Println("Starting ARRR payout")
						ok, result := coin.ARRR_Send(e.Destination, e.Price)
						log.Printf("ARRR status: %v, %s\n", ok, result)
						coin.LockedBalance.RemoveLockedBalance(e.Coin, e.Price)

						push := fmt.Sprintf("%s swap done", e.Coin)
						PushOver(push)
					case coin.DERO_XMR:
						xmr_txs = append(xmr_txs, monero.AddTX(e.Destination, e.Price))
					default:
						txs = append(txs, dero.AddTX(e.Destination, e.Amount))
					}

					e.Status = SWAP_DONE
					sent++
					active--
				} else {
					// transaction has been confirmed
					e.Status = SWAP_CONFIRMED
				}

				file_data, _ = json.Marshal(e)
				os.WriteFile(fmt.Sprintf("swaps/active/%d", t), file_data, 0644)

				if e.Status == SWAP_DONE {
					err = os.WriteFile(fmt.Sprintf("swaps/done/%d", t), file_data, 0644)
					if err != nil {
						log.Printf("Can't mark swap as done, swap %d, err %v\n", t, err)
					} else {
						delete(coin.ActiveSwaps, t)
						os.Remove(fmt.Sprintf("swaps/active/%d", t))
					}
				}
			}
		}

		// Dero and Monero payout process
		if len(txs) > 0 {
			log.Println("Starting DERO payout process")
			dero.Payout(txs)

			push := "Dero swap(s) done"
			PushOver(push)
		}
		// TODO: create function and TX verification
		if len(xmr_txs) > 0 {
			log.Println("Starting XMR payout process")
			if ok, txid := monero.XMRSend(xmr_txs); ok {
				log.Printf("XMR transaction (TXID %s) successfully sent\n", txid)

				push := "XMR swap(s) done"
				PushOver(push)
			} else {
				log.Println("Error sending XMR transaction")
			}
		}

		if sent+expired+fails > 0 {
			log.Printf("Swap processing: %d sent, %d expired, %d errors\n", sent, expired, fails)
		}
	}
}
