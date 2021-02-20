package pow

import (
	"github.com/dabankio/civil"
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
	AddrFrom string     `json:"addrFrom"`
	Date     civil.Date `json:"date"`
}

type UnlockBlock struct {
	UnlockBlockBase
	AddrTo   string          `json:"addrTo"`
	Balance  decimal.Decimal `json:"balance"`
	TimeSpan int64           `json:"timeSpan"`
	Height   uint64          `json:"height"`
}

type ReqUnlockedBlocks struct {
	ReqSign
	BalanceLst []UnlockBlock `json:"balanceLst"`
}
