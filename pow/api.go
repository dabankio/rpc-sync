package pow

import (
	"time"

	"github.com/shopspring/decimal"
)

type ReqSign struct {
	RequestSign string `json:"requestSign"`
	AppID       string `json:"appID"`
	SignPlain   string `json:"signPlain"`
	TimeSpan    string `json:"timeSpan"`
}

// signPlain = appID + ":" + timeSpan + ":" + signPlain;

type UnlockBlockBase struct {
	AddrFrom string    `json:"addrFrom"`
	Date     time.Time `json:"date"`
}

type UnlockBlock struct {
	UnlockBlockBase
	Id       int64           `json:"id"`
	AddrTo   string          `json:"addrTo"`
	Balance  decimal.Decimal `json:"balance"`
	TimeSpan int64           `json:"timeSpan"`
	Height   uint64          `json:"height"`
}

type ReqUnlockedBlocks struct {
	ReqSign
	BalanceLst []UnlockBlock `json:"balanceLst"`
}
