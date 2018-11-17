package muxservice

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/decred/dcrd/dcrutil"
)

// Ticket store basic information regarding PoS mining
type Ticket struct {
	OwnMempool   uint32  `json:"ownMempool"`
	Immature     uint32  `json:"immature"`
	Live         uint32  `json:"live"`
	TotalSubsidy float64 `json:"totalSubsidy"`
}

// BuyTicket using stake pool
func BuyTicket(spendLimit dcrutil.Amount, ticketAddress dcrutil.Address,
	numTickets *int, poolAddress dcrutil.Address, poolFees *dcrutil.Amount, ticketFee *dcrutil.Amount, unlockTimeout int64) (string, error) {
	hashResult := ""
	stakeInfo, err := client.GetStakeInfo()
	if err != nil {
		return "", err
	}
	logger.Tracef("Stake info:\n%v", spew.Sdump(stakeInfo))

	err = client.WalletPassphrase(cfg.WalletPass, unlockTimeout)
	if err != nil {
		return "", err
	}

	fromAccount := "default"
	expiry := 0
	minConf := 1
	splitTx := false
	hashes, err := client.PurchaseTicket(fromAccount, spendLimit, &minConf, ticketAddress, numTickets, poolAddress, poolFees, &expiry, &splitTx, ticketFee)
	for i := range hashes {
		logger.Infof("Purchased ticket %v at stake difficulty %v", hashes[i], stakeInfo.Difficulty)
		hashResult += hashes[i].String() + " | "
	}
	return hashResult, err
}

// GetTickets return Ticket object
func GetTickets() (Ticket, error) {
	var ticket Ticket

	stakeinfo, err := client.GetStakeInfo()
	if err != nil {
		logger.Errorf("Error %v", err)
		return ticket, err
	}

	logger.Tracef("StakeInfo:\n%v", spew.Sdump(stakeinfo))
	ticket.Immature = stakeinfo.Immature
	ticket.Live = stakeinfo.Live
	ticket.OwnMempool = stakeinfo.OwnMempoolTix
	ticket.TotalSubsidy = stakeinfo.TotalSubsidy

	return ticket, nil
}
