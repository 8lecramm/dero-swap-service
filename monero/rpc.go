package monero

// XMR related RPC structures
type (
	RPC_XMR_Height struct {
		Height uint64 `json:"height"`
	}
	RPC_XMR_GetBalance_Params struct {
		AccountIndex   uint64   `json:"account_index"`
		AddressIndices []uint64 `json:"address_indices,omitempty"`
		AllAccounts    bool     `json:"all_accounts"`
		Strict         bool     `json:"strict"`
	}
	RPC_XMR_GetBalance_Result struct {
		Balance         uint64                    `json:"balance"`
		UnlockedBalance uint64                    `json:"unlocked_balance"`
		TimeToUnlock    uint64                    `json:"time_to_unlock"`
		BlocksToUnlock  uint64                    `json:"blocks_to_unlock"`
		PerSubaddress   []RPC_XMR_Subaddress_Info `json:"per_subaddress"`
	}
	RPC_XMR_GetAddress_Params struct {
		AccountIndex uint64   `json:"account_index"`
		AddressIndex []uint64 `json:"address_index,omitempty"`
	}
	RPC_XMR_GetAddress_Result struct {
		Address   string              `json:"address"`
		Addresses []RPC_XMR_Addresses `json:"addresses"`
	}
	RPC_XMR_Addresses struct {
		Address      string `json:"address"`
		Label        string `json:"label"`
		AddressIndex uint64 `json:"address_index"`
		Used         bool   `json:"used"`
	}
	RPC_XMR_Subaddress_Info struct {
		AccountIndex      uint64   `json:"account_index"`
		AddressIndices    []uint64 `json:"address_indices,omitempty"`
		Address           string   `json:"address"`
		Balance           uint64   `json:"balance"`
		UnlockedBalance   uint64   `json:"unlocked_balance"`
		Label             string   `json:"label"`
		NumUnspentOutputs uint64   `json:"num_unspent_outputs"`
		TimeToUnlock      uint64   `json:"time_to_unlock"`
		BlocksToUnlock    uint64   `json:"blocks_to_unlock"`
	}
	RPC_XMR_IntegratedAddress_Result struct {
		IntegratedAddress string `json:"integrated_address"`
		PaymentID         string `json:"payment_id"`
	}
	RPC_XMR_SplitIntegratedAddress_Params struct {
		IntegratedAddress string `json:"integrated_address"`
	}
	RPC_XMR_SplitIntegratedAddress_Result struct {
		StandardAddress string `json:"standard_address"`
		PaymentID       string `json:"payment_id"`
		IsSubaddress    bool   `json:"is_subaddress"`
	}
	RPC_XMR_Validate_Address_Params struct {
		Address         string `json:"address"`
		AnyNetType      bool   `json:"any_net_type"`
		Allow_Openalias bool   `json:"allow_openalias"`
	}
	RPC_XMR_Validate_Address_Result struct {
		Valid            bool   `json:"valid"`
		Integrated       bool   `json:"integrated"`
		Subaddress       bool   `json:"subaddress"`
		Nettype          string `json:"nettype"`
		OpenaliasAddress string `json:"openalias_address"`
	}
	RPC_XMR_Transfer_Params struct {
		Address string `json:"address"`
		Amount  uint64 `json:"amount"`
	}
	RPC_XMR_Transfer struct {
		Destinations   []RPC_XMR_Transfer_Params `json:"destinations"`
		AccountIndex   uint64                    `json:"account_index"`
		SubaddrIndices []uint64                  `json:"subaddr_indices"`
		Ringsize       uint64                    `json:"ring_size"`
		UnlockTime     uint64                    `json:"unlock_time"`
		PaymentID      string                    `json:"payment_id"`
		GetTXKey       bool                      `json:"get_tx_keys"`
		Priority       uint64                    `json:"priority"`
		DoNotRelay     bool                      `json:"do_not_relay"`
		GetTXHash      bool                      `json:"get_tx_hash"`
		GetTxMetadata  bool                      `json:"get_tx_metadata"`
	}
	RPC_XMR_Transfer_Result struct {
		Amount         uint64 `json:"amount"`
		Fee            uint64 `json:"fee"`
		MultiSig_TxSet string `json:"multisig_txset"`
		TxBlob         string `json:"tx_blob"`
		TxHash         string `json:"tx_hash"`
		TxKey          string `json:"tx_key"`
		TxMetadata     string `json:"tx_metadata"`
		Unsigned_TxSet string `json:"unsigned_txset"`
	}
	RPC_XMR_Payments struct {
		Address     string `json:"address"`
		Amount      uint64 `json:"amount"`
		BlockHeight uint64 `json:"block_height"`
		PaymentID   string `json:"payment_id"`
		TxHash      string `json:"tx_hash"`
		UnlockTime  uint64 `json:"unlock_time"`
		Locked      bool   `json:"locked"`
	}
	RPC_XMR_GetPayments_Result struct {
		Payments []RPC_XMR_Payments `json:"payments"`
	}
	RPC_XMR_BulkTX_Params struct {
		Payment_IDs    []string `json:"payment_ids"`
		MinBlockHeight uint64   `json:"min_block_height"`
	}
)
