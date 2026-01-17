package utils

import (
	"fmt"
	"strings"
	"time"
)

func GnerateOutTradeNo(mchID string) string {
	timestamp := time.Now().Format("060102150405.000")
	timestamp = strings.Replace(timestamp, ".", "", 1)
	timestamp = timestamp[:14]

	if len(mchID) > 4 {
		mchID = mchID[len(mchID)-4:]
	}

	randStr := RandomString(6)

	return fmt.Sprintf("%s%s%s", timestamp, mchID, randStr)
}
