package pow

import (
	"github.com/dabankio/civil"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

func NewRepo(db *sqlx.DB) *Repo { return &Repo{DB: db} }

type Repo struct {
	DB *sqlx.DB
}

type UnlockedBlock struct {
	AddrFrom  string
	AddrTo    string
	Balance   decimal.Decimal
	TimeSpan  int64
	Day       civil.Date
	RewardDay civil.Date
	Height    uint64
}

func (r *Repo) InsertUnlockedBlocks(blocks []UnlockedBlock, dbTx *sqlx.Tx) error {
	// _, err := r.db.NamedExec(`insert into unlocked_block (addr_from, addr_to, balance,time_span,day, height)
	// values (:addr_from, :addr_to, :balance, :time_span, :day, :height)`, blocks) //这个sql是可行的，但批量插入不能做on confilict do update

	for _, ub := range blocks {
		//风险提示：on confilict do update 可能存在一种情况导致数据错误：旧数据需要被删除（比如之前有 height_from_to 的记录，后续重新写入没有这个记录但这个旧记录没被删除）
		_, err := dbTx.Exec(`insert into unlocked_block (addr_from, addr_to, balance, time_span, day, reward_day, height) 
		values ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ON CONSTRAINT unq_height_from_to DO UPDATE 
		SET balance = $3, time_span = $4, day = $5, reward_day = $6`, ub.AddrFrom, ub.AddrTo, ub.Balance, ub.TimeSpan, ub.Day, ub.RewardDay, ub.Height)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) QueryUnlockedBlocks(addrFrom string, rewardDay civil.Date) (items []UnlockBlock, err error) {
	err = r.DB.Select(`select * from unlocked_block where addr_from = $1 and reward_day = $2`, addrFrom, rewardDay)
	return
}
