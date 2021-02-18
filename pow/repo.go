package pow

import (
	"github.com/dabankio/civil"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

type Repo struct {
	db *sqlx.DB
}

type UnlockedBlock struct {
	AddrFrom string
	AddrTo   string
	Balance  decimal.Decimal
	TimeSpan int64
	Day      civil.Date
	Height   uint64
}

func (r *Repo) InsertUnlockedBlocks(blocks []UnlockedBlock) error {
	_, err := r.db.NamedExec(`insert into unblocked_block (addr_from, addr_to, balance,time_span,day, height) 
	values (:addr_from, :addr_to, :balance, :time_span, :day, :height)`, blocks)
	return err
}

func (r *Repo) QueryUnlockedBlocks(addrFrom string, day civil.Date) (items []UnlockBlock, err error) {
	err = r.db.Select(`select * from unblocked_block where addr_from = $1 and day = $2`, addrFrom, day)
	return
}
