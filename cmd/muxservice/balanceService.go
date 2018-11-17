package muxservice

import "github.com/davecgh/go-spew/spew"

// Balance stores total amount for accounts
type Balance struct {
	Amount float64 `json:"amount"`
}

// GetBalance calculates and returns the balance of all accounts
func GetBalance() (Balance, error) {
	var balance Balance
	balance.Amount = 0

	balances, err := client.GetBalance("*")
	if err != nil {
		logger.Errorf("Error %v", err)
		return balance, err
	}

	if len(balances.Balances) > 0 {
		for _, value := range balances.Balances {
			logger.Tracef("Balance:\n%v", spew.Sdump(value))
			balance.Amount += value.Spendable
		}
	}

	return balance, nil
}
