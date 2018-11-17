package muxservice

import (
	"crypto"
	"errors"
	"io/ioutil"
	"os"

	"github.com/sec51/twofactor"
)

// GetTwoFactorQR generate a byte array for account and issuer
func GetTwoFactorQR(account string, issuer string, otpFileName string) ([]byte, error) {
	if _, err := os.Stat(otpFileName); err == nil { // Check if file exists
		return nil, errors.New("OTP: Two factor authentication already set")
	}
	otp, err := twofactor.NewTOTP(account, issuer, crypto.SHA1, 6)
	if err != nil {
		return nil, err
	}
	qrBytes, err := otp.QR()
	if err != nil {
		return nil, err
	}
	otpBytes, err := otp.ToBytes()
	err = ioutil.WriteFile(otpFileName, otpBytes, 0644)
	if err != nil {
		return nil, err
	}

	return qrBytes, nil
}
