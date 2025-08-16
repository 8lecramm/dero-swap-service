package cfg

import (
	"dero-swap/coin"
	"dero-swap/dero"
	"dero-swap/monero"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/ybbus/jsonrpc/v3"
)

func isServer() bool {
	if len(os.Args) == 2 && os.Args[1] == "--server" {
		Settings.Mode = SERVER // Set mode to server
		return true
	}
	return false
}

// load config file
func LoadConfig() {

	fd, err := os.ReadFile("config.json")
	if err != nil {
		log.Printf("Error loading config file: %v\n", err)
		return
	}
	err = json.Unmarshal(fd, &Settings)
	if err != nil {
		log.Printf("Error parsing config file: %v\n", err)
		return
	}

	// Setup RPC for Dero Daemon and Wallet
	dero.Dero_Daemon = jsonrpc.NewClient("http://" + Settings.Dero_Daemon + "/json_rpc")
	if Settings.Dero_Login != "" {
		dero.RPC_Login = base64.StdEncoding.EncodeToString([]byte(Settings.Dero_Login))
		dero.Dero_Wallet = jsonrpc.NewClientWithOpts("http://"+Settings.Dero_Wallet+"/json_rpc", &jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + dero.RPC_Login,
			},
		})
	} else {
		dero.Dero_Wallet = jsonrpc.NewClient("http://" + Settings.Dero_Wallet + "/json_rpc")
		log.Println("Dero Wallet: No RPC authorization specified")
	}

	// Setup RPC for Monero Wallet
	monero.Monero_Wallet = jsonrpc.NewClient("http://" + Settings.Monero_Wallet + "/json_rpc")

	// Setup RPC for Bitcoin-like coins
	coin.BTC_Login = Settings.BTC_Login
	coin.LTC_Login = Settings.LTC_Login
	coin.XTC_URL[coin.BTC] = "http://" + Settings.BTC_Daemon
	coin.XTC_URL[coin.LTC] = "http://" + Settings.LTC_Daemon
	coin.XTC_URL[coin.ARRR] = "http://" + Settings.ARRR_Daemon

	// check if pair is "supported"
	for _, p := range Settings.Pairs {
		supported := false
		for i := range SupportedCoins {
			if p == SupportedCoins[i] {
				supported = true
				break
			}
		}
		if supported {
			coin.Pairs[p] = true
		} else {
			log.Printf("%s is not a supported pair\n", p)
		}
	}

	log.Printf("Config successfully loaded\n")

	LoadFees()
}

// load fees file
func LoadFees() {

	fd, err := os.ReadFile("fees.json")
	if err != nil {
		log.Printf("Error loading fees file: %v\n", err)
		return
	}
	err = json.Unmarshal(fd, &SwapFees)
	if err != nil {
		log.Printf("Error parsing fees file: %v\n", err)
		return
	}

	if Settings.Mode == SERVER {
		log.Printf("%-14s: %.2f%%\n", "Fees", SwapFees.Fees)
	}
}

// basic config check
func CheckConfig() bool {

	if !isServer() {
		if Settings.ServerAddress == "" || Settings.Nickname == "" {
			log.Println("Server address and/or nickname are required when server address is set")
			return false
		}
		Settings.ServerURL = url.URL{Scheme: "wss", Host: Settings.ServerAddress, Path: "/ws"}

		// make all pairs available
		for p := range coin.Pairs {
			coin.IsPairAvailable[p] = true
		}
	}

	if Settings.Dero_Daemon == "" || Settings.Dero_Wallet == "" {
		log.Println("Dero Daemon or Dero Wallet is not set")
		return false
	}

	for p := range coin.Pairs {
		switch p {
		case BTC:
			if Settings.BTC_Daemon == "" || Settings.BTC_Dir == "" {
				log.Printf("%s pair is set, but daemon or directory is not set\n", p)
				return false
			} else {
				coin.BTC_Dir = Settings.BTC_Dir
			}
		case LTC:
			if Settings.LTC_Daemon == "" || Settings.LTC_Dir == "" {
				log.Printf("%s pair is set, but daemon or directory is not set\n", p)
				return false
			} else {
				coin.LTC_Dir = Settings.LTC_Dir
			}
		case ARRR:
			if Settings.ARRR_Daemon == "" || Settings.ARRR_Dir == "" {
				log.Printf("%s pair is set, but daemon or directory is not set\n", p)
				return false
			} else {
				coin.ARRR_Dir = Settings.ARRR_Dir
			}
		case XMR:
			if Settings.Monero_Wallet == "" {
				log.Printf("%s pair is set, but wallet is not set\n", p)
				return false
			}
		}
	}

	if dero.RPC_Login != "" {
		log.Printf("%-14s: %s\n", "Dero Wallet", "Using RPC basic auth")
	}
	wallet := dero.GetAddress()

	if dero.GetHeight() == 0 || dero.CheckBlockHeight() == 0 || wallet == "" {
		log.Println("Dero daemon or wallet is not available")
		return false
	}
	log.Printf("%-14s: %s\n", "Dero Wallet", wallet)

	return true
}

// TODO: optimization needed
func LoadWallets() {

	for p := range coin.Pairs {
		switch p {
		case ARRR:
			addr := coin.ARRR_GetAddress()
			if !coin.XTCValidateAddress(p, addr) {
				log.Printf("Disable pair \"%s\": wallet not available or other error\n", p)
				delete(coin.Pairs, p)
			} else {
				if coin.ARRR_address == "" {
					coin.ARRR_address = addr
					log.Printf("%-14s: %s\n", "ARRR Wallet", addr)
				}
			}
		case XMR:
			addr := monero.GetAddress()
			if !monero.ValidateAddress(addr) {
				log.Printf("Disable pair \"%s\": wallet not available or other error\n", p)
				delete(coin.Pairs, p)
			} else {
				if coin.XMR_address == "" {
					coin.XMR_address = addr
					log.Printf("%-14s: %s\n", "XMR Wallet", addr)
				}
			}
		case LTC:
			ok, err := coin.XTCLoadWallet(p)
			if !ok && !strings.Contains(err, "is already loaded") {
				ok, err := coin.XTCNewWallet(p)
				if !ok {
					log.Printf("Disable pair \"%s\": %s\n", p, err)
					delete(coin.Pairs, p)
				} else {
					addr := coin.XTCGetAddress(p)
					if coin.LTC_address == "" {
						coin.LTC_address = addr
						log.Printf("%-14s: %s\n", "LTC Wallet", addr)
					}
				}
			} else {
				addr := coin.XTCGetAddress(p)
				if coin.LTC_address == "" {
					coin.LTC_address = addr
					log.Printf("%-14s: %s\n", "LTC Wallet", addr)
				}
			}
		case BTC:
			ok, err := coin.XTCLoadWallet(p)
			// TODO:
			// sometimes the wallet needs to much time to load
			// skipping wallet creation for now
			/*
				if !ok && !strings.Contains(err, "is already loaded") {
						ok, err := coin.XTCNewWallet(p)
						if !ok {
							log.Printf("Disable pair \"%s\": %s\n", p, err)
							delete(Pairs, p)
						} else {
							addr := coin.XTCGetAddress(p)
							if BTC_address == "" {
								BTC_address = addr
								log.Printf("%-14s: %s\n", "BTC Wallet", addr)
							}
						}
			*/
			if !ok && !strings.Contains(err, "is already loaded") {
				log.Printf("Disable pair \"%s\": %s\n", p, err)
				delete(coin.Pairs, p)
			} else {
				addr := coin.XTCGetAddress(p)
				if coin.BTC_address == "" {
					coin.BTC_address = addr
					log.Printf("%-14s: %s\n", "BTC Wallet", addr)
				}
			}
		default:
			continue
		}
	}
}

func SetFieldByJSONTag(s interface{}, tag string, value interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("must pass a pointer to a struct")
	}
	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("must pass a pointer to a struct")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		jsonTag, _ := strings.CutSuffix(structField.Tag.Get("json"), ",omitempty") // remove ",omitempty" if present

		if jsonTag == tag {
			field := v.Field(i)
			if !field.CanSet() {
				return fmt.Errorf("cannot set field %s", structField.Name)
			}
			val := reflect.ValueOf(value)
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
			} else if val.Type().ConvertibleTo(field.Type()) {
				field.Set(val.Convert(field.Type()))
			} else {
				return fmt.Errorf("cannot assign value of type %s to field %s of type %s", val.Type(), structField.Name, field.Type())
			}
			return nil
		}
	}
	return fmt.Errorf("no field with JSON tag %q found", tag)
}
