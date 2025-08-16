package coin

// RPC request / response
type (
	RPC_Request struct {
		Jsonrpc string        `json:"jsonrpc"`
		Id      string        `json:"id"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
	}
	RPC_Response struct {
		Error RPC_Error `json:"error"`
	}
	RPC_Send_Result struct {
		Result string `json:"result"`
	}
	RPC_Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// BTC/LTC related RPC structures
type (
	RPC_NewWallet_Response struct {
		Result RPC_NewWallet `json:"result"`
		Error  RPC_Error     `json:"error"`
	}
	RPC_NewWallet struct {
		Name    string `json:"name"`
		Warning string `json:"warning"`
	}
	RPC_NewAddress_Response struct {
		Result string    `json:"result"`
		Error  RPC_Error `json:"error"`
	}
	RPC_GetBlockCount_Response struct {
		Result uint64    `json:"result"`
		Error  RPC_Error `json:"error"`
	}
	RPC_ReceivedByAddress_Response struct {
		Result float64   `json:"result"`
		Error  RPC_Error `json:"error"`
	}
	RPC_ListReceivedByAddress_Response struct {
		Result []RPC_ListReceivedByAddress `json:"result"`
	}
	RPC_ListReceivedByAddress struct {
		InvolvesWatchonly bool     `json:"involvesWatchonly"`
		Address           string   `json:"address"`
		Amount            float64  `json:"amount"`
		Confirmations     uint64   `json:"confirmations"`
		Label             string   `json:"label"`
		Txids             []string `json:"txids"`
	}
	RPC_GetTransaction_Response struct {
		Result RPC_GetTransaction `json:"result"`
	}
	RPC_GetTransaction struct {
		Amount            float64       `json:"amount"`
		Fee               float64       `json:"fee"`
		Confirmations     uint64        `json:"confirmations"`
		Generated         bool          `json:"generated"`
		Trusted           bool          `json:"trusted"`
		Blockhash         string        `json:"blockhash"`
		Blockheight       uint64        `json:"blockheight"`
		Blockindex        uint64        `json:"blockindex"`
		Blocktime         uint64        `json:"blocktime"`
		Txid              string        `json:"txid"`
		Walletconflicts   []string      `json:"walletconflicts"`
		Time              uint64        `json:"time"`
		TimeReceived      uint64        `json:"timereceived"`
		Comment           string        `json:"comment"`
		Bip125Replaceable string        `json:"bip125-replaceable"`
		Detail            []RPC_Details `json:"details"`
		Hex               string        `json:"hex"`
	}
	RPC_GetBalance_Result struct {
		Result float64 `json:"result"`
	}
	RPC_Details struct {
		InvolvesWatchonly bool    `json:"involvesWatchonly"`
		Address           string  `json:"address"`
		Category          string  `json:"category"`
		Amount            float64 `json:"amount"`
		Label             string  `json:"label"`
		Vout              float64 `json:"vout"`
		Fee               float64 `json:"fee"`
		Abandoned         bool    `json:"abandoned"`
	}
	RPC_Validate_Address struct {
		IsValid         bool    `json:"isvalid"`
		Address         string  `json:"address"`
		ScriptPubKey    string  `json:"scriptPubKey"`
		Isscript        bool    `json:"isscript"`
		Iswitness       bool    `json:"iswitness"`
		Witness_version float64 `json:"witness_version,omitempty"`
		Witness_program string  `json:"witness_program,omitempty"`
	}
	RPC_Validate_Address_Result struct {
		Result RPC_Validate_Address `json:"result"`
	}
)

// ARRR related RPC structures
type (
	RPC_ARRR_ListReceivedByAddress_Response struct {
		Result []RPC_ARRR_ListReceivedByAddress `json:"result"`
	}
	RPC_ARRR_ListReceivedByAddress_Details struct {
		Type      string  `json:"type"`
		Output    uint64  `json:"output"`
		Outgoing  bool    `json:"outgoing"`
		Address   string  `json:"address"`
		Value     float64 `json:"value"`
		ValueZat  uint64  `json:"valueZat"`
		Change    bool    `json:"change"`
		Spendable bool    `json:"spendable"`
		Memo      string  `json:"memo"`
		MemoStr   string  `json:"memoStr"`
	}
	RPC_ARRR_ListAddresses struct {
		Result []string `json:"result"`
	}
	RPC_ARRR_ListReceivedByAddress struct {
		Txid             string                                   `json:"txid"`
		Coinbase         bool                                     `json:"coinbase"`
		Category         string                                   `json:"category"`
		BlockHeight      uint64                                   `json:"blockHeight"`
		Blockhash        string                                   `json:"blockhash"`
		Blockindex       uint64                                   `json:"blockindex"`
		Blocktime        uint64                                   `json:"blocktime"`
		RawConfirmations uint64                                   `json:"rawconfirmations"`
		Confirmations    uint64                                   `json:"confirmations"`
		Time             uint64                                   `json:"time"`
		ExpiryHeight     uint64                                   `json:"expiryHeight"`
		Size             uint64                                   `json:"size"`
		Fee              uint64                                   `json:"fee"`
		Received         []RPC_ARRR_ListReceivedByAddress_Details `json:"received"`
	}
	RPC_ARRR_SendMany_Params struct {
		Address string  `json:"address"`
		Amount  float64 `json:"amount"`
		Memo    string  `json:"memo,omitempty"`
	}
)
