package coin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"net/http"
)

// TODO: refactoring needed

func XTCGetCookie(pair string) (bool, string) {

	var data []byte
	var err error

	switch pair {
	case "btc":
		if data, err = os.ReadFile(BTC_Dir + "/.cookie"); err != nil {
			log.Printf("Can't load BTC auth cookie: %v\n", err)
			return false, ""
		}
	case "ltc":
		if data, err = os.ReadFile(LTC_Dir + "/.cookie"); err != nil {
			log.Printf("Can't load LTC auth cookie: %v\n", err)
			return false, ""
		}
	case "arrr":
		if data, err = os.ReadFile(ARRR_Dir + "/.cookie"); err != nil {
			log.Printf("Can't load ARRR auth cookie: %v\n", err)
			return false, ""
		}
	}

	auth := base64.StdEncoding.EncodeToString(data)
	data = nil

	return true, auth
}

func SetHeaders(request *http.Request) {
	request.Header.Add("Content-Type", "text/plain")
}

func CheckRPCAuth(pair string) (ok bool, auth string) {

	var credentials string

	pair = GetPair(pair)

	switch pair {
	case "btc":
		if BTC_Login != "" {
			credentials = base64.StdEncoding.EncodeToString([]byte(BTC_Login))
			ok = true
		}
	case "ltc":
		if LTC_Login != "" {
			credentials = base64.StdEncoding.EncodeToString([]byte(LTC_Login))
			ok = true
		}
	}

	if !ok {
		ok, credentials = XTCGetCookie(pair)
		auth = "Basic " + credentials
	} else {
		auth = "Basic " + credentials
	}

	return
}

func XTCBuildRequest(pair string, method string, options []any) (*http.Request, error) {

	json_object := &RPC_Request{
		Jsonrpc: "1.0",
		Id:      "swap",
		Method:  method,
		Params:  options,
	}
	json_bytes, err := json.Marshal(&json_object)
	if err != nil {
		log.Printf("Can't marshal %s request: %v\n", method, err)
		return nil, err
	}

	req, err := http.NewRequest("POST", XTC_GetURL(pair), bytes.NewBuffer(json_bytes))
	if err != nil {
		log.Printf("Can't create %s request: %v\n", method, err)
		return nil, err
	}

	if ok, auth := CheckRPCAuth(pair); ok {
		if auth != "" {
			req.Header.Add("Authorization", auth)
		}
	} else {
		log.Println("Invalid RPC auth")
	}

	SetHeaders(req)

	return req, nil
}

func XTCSendRequest(request *http.Request, response any) error {

	resp, err := XTC_Daemon.Do(request)
	if err != nil {
		log.Printf("Can't send %s request: %v\n", request.Method, err)
		return err
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, response)
	if err != nil {
		log.Printf("Can't unmarshal %s response: %v\n", request.Method, err)
		return err
	}

	return nil
}

func XTCNewWallet(pair string) (result bool, message string) {

	switch pair {
	case DERO_ARRR, ARRR, DERO_XMR, XMR:
		return false, ""
	}
	req, err := XTCBuildRequest(pair, "createwallet", []any{"swap_wallet"})
	if err != nil {
		log.Printf("Can't build createwallet request: %v\n", err)
		return false, ""
	}

	var json_resonse RPC_NewWallet_Response

	err = XTCSendRequest(req, &json_resonse)
	if err != nil {
		return false, err.Error()
	}

	if json_resonse.Error != (RPC_Error{}) {
		message = json_resonse.Error.Message
	} else {
		result = true
	}

	return result, message
}

func XTCLoadWallet(pair string) (result bool, message string) {

	req, err := XTCBuildRequest(pair, "loadwallet", []any{"swap_wallet"})
	if err != nil {
		log.Printf("Can't build loadwallet request: %v\n", err)
		return false, ""
	}

	var json_resonse RPC_NewWallet_Response

	err = XTCSendRequest(req, &json_resonse)
	if err != nil {
		return false, err.Error()
	}

	if json_resonse.Error != (RPC_Error{}) {
		message = json_resonse.Error.Message
	} else {
		result = true
	}

	return result, message
}

func XTCNewAddress(pair string) string {

	req, err := XTCBuildRequest(pair, "getnewaddress", []any{})
	if err != nil {
		log.Printf("Can't build getnewaddress request: %v\n", err)
		return ""
	}

	var json_response RPC_NewAddress_Response

	err = XTCSendRequest(req, &json_response)
	if err != nil {
		return err.Error()
	}
	log.Println("Successfully created new address")

	return json_response.Result
}

func XTCGetAddress(pair string) string {

	switch pair {
	case DERO_ARRR, ARRR, DERO_XMR, XMR:
		return ""
	}

	_, _, address, _ := XTCListReceivedByAddress(pair, "", 0, 0, true)

	return address
}

func XTCCheckBlockHeight(pair string) uint64 {

	req, err := XTCBuildRequest(pair, "getblockcount", []any{})
	if err != nil {
		log.Printf("Can't build getblockcount request: %v\n", err)
		return 0
	}

	var response RPC_GetBlockCount_Response

	err = XTCSendRequest(req, &response)
	if err != nil {
		return 0
	} else {
		return response.Result
	}
}

func XTCReceivedByAddress(pair string, wallet string) (float64, error) {

	req, err := XTCBuildRequest(pair, "getreceivedbyaddress", []any{wallet, 2})
	if err != nil {
		log.Printf("Can't build getreceivedbyaddress request: %v\n", err)
		return 0, err
	}

	var response RPC_ReceivedByAddress_Response

	err = XTCSendRequest(req, &response)
	if err != nil {
		return 0, err
	} else {
		return response.Result, nil
	}
}

func XTCListReceivedByAddress(pair string, wallet string, amount float64, height uint64, get_address bool) (bool, bool, string, error) {

	var method string
	var options []any

	if !get_address {
		if pair == ARRR {
			method = "zs_listreceivedbyaddress"
			options = append(options, wallet, 1, 3, height)
		} else {
			method = "listreceivedbyaddress"
			options = append(options, 1, false, false, wallet)
		}
	} else {
		method = "listreceivedbyaddress"
		options = append(options, 0, true)
	}

	req, err := XTCBuildRequest(pair, method, options)
	if err != nil {
		log.Printf("Can't build listreceivedbyaddress request: %v\n", err)
		return false, false, "", err
	}

	if pair != ARRR {
		var response RPC_ListReceivedByAddress_Response

		err = XTCSendRequest(req, &response)
		if err != nil {
			return false, false, "", err
		}

		if !get_address {
			if len(response.Result) > 0 {
				for _, tx := range response.Result[0].Txids {
					tx_data, err := XTCGetTransaction(pair, tx)
					if err != nil {
						continue
					}
					if tx_data.Result.Amount == amount && tx_data.Result.Blockheight >= height {
						if tx_data.Result.Confirmations > 1 {
							return true, true, "", nil
						} else {
							return false, true, "", nil
						}
					}
				}
			}
		} else {
			if len(response.Result) > 0 {
				return false, false, response.Result[0].Address, nil
			} else {
				return false, false, "", nil
			}
		}
	} else {
		var response RPC_ARRR_ListReceivedByAddress_Response

		err = XTCSendRequest(req, &response)
		if err != nil {
			return false, false, "", err
		}

		if len(response.Result) > 0 {
			for e := range response.Result {
				if response.Result[e].BlockHeight >= height {
					for _, tx := range response.Result[e].Received {
						if tx.Value == amount {
							if response.Result[e].Confirmations > 1 {
								return true, true, "", nil
							} else {
								return false, true, "", nil
							}
						}
					}
				}
			}
		}
	}

	return false, false, "", nil
}

func XTCGetTransaction(pair string, txid string) (result RPC_GetTransaction_Response, err error) {

	req, err := XTCBuildRequest(pair, "gettransaction", []any{txid, false})
	if err != nil {
		log.Printf("Can't build gettransaction request: %v\n", err)
		return result, err
	}

	err = XTCSendRequest(req, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func XTCGetBalance(pair string) float64 {

	var method string
	var options []any
	if pair == DERO_ARRR || pair == ARRR {
		method = "z_getbalance"
		options = append(options, ARRR_GetAddress())
	} else {
		method = "getbalance"
	}

	req, err := XTCBuildRequest(pair, method, options)
	if err != nil {
		log.Printf("Can't build getbalance request: %v\n", err)
		return 0
	}

	var result RPC_GetBalance_Result

	err = XTCSendRequest(req, &result)
	if err != nil {
		return 0
	}

	if strings.HasPrefix(pair, "dero") {
		result.Result -= LockedBalance.GetLockedBalance(pair)
	}

	return RoundFloat(result.Result, 8)
}

func XTCSend(pair string, wallet string, amount float64, fee uint64) (bool, string) {

	var options []any

	switch pair {
	case DERO_LTC:
		options = append(options, wallet, amount)
	case DERO_BTC:
		options = append(options, wallet, amount, "", "", false, true, nil, "unset", nil, fee)
	}

	req, err := XTCBuildRequest(pair, "sendtoaddress", options)
	if err != nil {
		log.Printf("Can't build sendtoaddress request: %v\n", err)
		return false, ""
	}

	var result RPC_Send_Result

	err = XTCSendRequest(req, &result)
	if err != nil {
		return false, ""
	}

	return true, result.Result
}

func XTCValidateAddress(pair string, address string) bool {

	var method string
	if pair == DERO_ARRR || pair == ARRR {
		method = "z_validateaddress"
	} else {
		method = "validateaddress"
	}

	req, err := XTCBuildRequest(pair, method, []any{address})
	if err != nil {
		log.Printf("Can't build validateaddress request: %v\n", err)
		return false
	}

	var result RPC_Validate_Address_Result

	err = XTCSendRequest(req, &result)
	if err != nil {
		return false
	}

	return result.Result.IsValid
}

func XTC_GetURL(pair string) string {

	pair = GetPair(pair)

	switch pair {
	case "btc":
		return XTC_URL[BTC]
	case "ltc":
		return XTC_URL[LTC]
	case "arrr":
		return XTC_URL[ARRR]
	default:
		return ""
	}
}

func ARRR_Send(wallet string, amount float64) (bool, string) {

	var params []RPC_ARRR_SendMany_Params

	params = append(params,
		RPC_ARRR_SendMany_Params{
			Address: wallet,
			Amount:  amount,
		})
	options := []any{ARRR_GetAddress(), params}

	req, err := XTCBuildRequest(DERO_ARRR, "z_sendmany", options)
	if err != nil {
		log.Printf("Can't build z_sendmany request: %v\n", err)
		return false, ""
	}

	var result RPC_Send_Result

	err = XTCSendRequest(req, &result)
	if err != nil {
		return false, ""
	}

	return true, result.Result
}

func ARRR_GetAddress() string {

	req, err := XTCBuildRequest(DERO_ARRR, "z_listaddresses", nil)
	if err != nil {
		log.Printf("Can't build z_listaddresses request: %v\n", err)
		return ""
	}

	var result RPC_ARRR_ListAddresses

	err = XTCSendRequest(req, &result)
	if err != nil {
		return ""
	}

	if len(result.Result) == 0 {
		return ""
	} else {
		return result.Result[0]
	}
}
