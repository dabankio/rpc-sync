package reward

import (
	"github.com/dabankio/civil"
	"github.com/shopspring/decimal"
)

type Reward struct {
	// FromTime   time.Time
	// ToTime     time.Time
	FromHeight uint64
	ToHeight   uint64
	// BlockCount uint

	Delegate string
	Voter    string
	Amount   decimal.Decimal
}

type DayReward struct {
	Day      civil.Date
	Delegate string
	Voter    string
	Amount   decimal.Decimal
}
