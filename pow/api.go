package pow

import "github.com/shopspring/decimal"

type ReqSign struct {
	RequestSign string `json:"requestSign"`
	AppID       string `json:"appID"`
	SignPlain   string `json:"signPlain"`
	TimeSpan    string `json:"timeSpan"`
}

// signPlain = appID + ":" + timeSpan + ":" + signPlain;

type UnlockBlockBase struct {
	AddrFrom string `json:"addrFrom"`
	Date     string `json:"date"`
}

type UnlockBlock struct {
	UnlockBlockBase
	Id       int64           `json:"id"`
	AddrTo   string          `json:"addrTo"`
	Balance  decimal.Decimal `json:"balance"`
	TimeSpan int64           `json:"timeSpan"`
	Height   int             `json:"height"`
}

type ReqUnblocks struct {
	ReqSign
	BalanceLst []UnlockBlock `json:"balanceLst"`
}
