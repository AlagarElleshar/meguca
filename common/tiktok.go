package common

import (
	"strconv"
	"time"
)

func GetTokID(filename string) *string {
	digits := ""
	for _, c := range filename {
		if c >= '0' && c <= '9' {
			digits += string(c)
		} else {
			if IsValidTokID(digits) {
				return &digits
			}
			digits = ""
		}
	}
	if IsValidTokID(digits) {
		return &digits
	}
	return nil
}

// This function validates tok IDs
func IsValidTokID(digits string) bool {
	//Max TokID is a week from now
	//Min TokID is 2016-08-01
	now := time.Now()
	maxTokID := now.Add(time.Hour*24*7).Unix() << 32
	var minTokID int64 = 6313705004335104000
	numDigits := len(digits)
	if numDigits > 20 || numDigits < 19 {
		return false
	}
	tokID, _ := strconv.ParseInt(digits, 10, 64)
	return tokID > minTokID && tokID < maxTokID
}
