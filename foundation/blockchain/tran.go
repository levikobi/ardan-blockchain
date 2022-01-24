package blockchain

import (
	"github.com/google/uuid"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// Set of transaction status types.
const (
	TxStatusAccepted = "accepted" // Accepted and should be applied to the balance.
	TxStatusError    = "error"    // An error occured and should not be applied to the balance.
	TxStatusNew      = "new"      // The transaction is newly added to the mempool.
)

// =============================================================================

// TxError represents an error on a transaction.
type TxError struct {
	Tx  Tx
	Err error
}

// Error implements the error interface.
func (txe *TxError) Error() string {
	return txe.Err.Error()
}

// =============================================================================

// ID represents a unique ID in the system.
type ID string

// Tx represents a transaction in the block.
type Tx struct {
	ID         ID     `json:"id"`          // Unique ID for the transaction to help with mempool lookups.
	From       string `json:"from"`        // The account this transaction is from.
	To         string `json:"to"`          // The account receiving the benefit of the transaction.
	Value      uint   `json:"value"`       // The monetary value received from this transactions.
	Data       string `json:"data"`        // Extra data related to the transaction.
	GasPrice   uint   `json:"gas_price"`   // The actual amount of gas spent to execute the transaction.
	GasLimit   uint   `json:"gas_limit"`   // The max amount of gas associated with the transaction.
	Status     string `json:"status"`      // The final status of the transaction to help reconcile balances.
	StatusInfo string `json:"status_info"` // Extra information related to the state.
}

// NewTx constructs a new TxRecord.
func NewTx(from string, to string, value uint, data string) Tx {
	return Tx{
		ID:     ID(uuid.New().String()),
		From:   from,
		To:     to,
		Value:  value,
		Data:   data,
		Status: TxStatusNew,
	}
}