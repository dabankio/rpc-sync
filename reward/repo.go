package reward

import (
	"bbcsyncer/infra"
	"time"

	"github.com/dabankio/civil"
	"github.com/jmoiron/sqlx"
)

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

type Repo struct {
	db *sqlx.DB
}

// 某一高度的投票情况汇总
func (r *Repo) CreateVoteSumAtHeight(height uint64) error {
	_, err := r.db.Exec(`
insert into vote_sum (block_height, delegate, voter, last_height, amount)
select $1 block_height, delegate, voter, max(block_height) last_height, sum(amount) amount
from dpos_vote
where block_height <= $1
group by delegate, voter`, height)
	return err
}

// 查询截止某高度的投票汇总
func (r *Repo) SelectVoteSumAtHeight(height uint64) ([]VoteSum, error) {
	var items []VoteSum
	err := r.db.Select(&items, `select * from vote_sum where block_height = $1`, height)
	return items, err
}

// BlockHeightBetween 时间范围(含端点)对应的区块高度
func (r *Repo) BlockHeightBetween(fromTime, toTime time.Time) (fromHeight, toHeight uint64, err error) {
	type endpointHeight struct{ Lower, Upper uint64 }
	var endpoint endpointHeight
	err = r.db.Get(&endpoint,
		`select min(height) as lower, max(height) as upper from blocks where time >= $1 and time <= $2`, fromTime, toTime)
	return endpoint.Lower, endpoint.Upper, err
}

func (r *Repo) InsertDayReward(items []DayReward) error {
	for _, itm := range items {
		_, err := r.db.Exec(`insert into day_reward (day, delegate,voter, amount) values ($1, $2, $3, $4)`,
			itm.Day, itm.Delegate, itm.Voter, itm.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) MaxDayOfDayReward() (*civil.Date, error) {
	var ret civil.Date
	err := r.db.Get(&ret, `select max(day)::text as day from day_reward`)
	return &ret, infra.WrapErr(err, "get max day of daily reward failed")
}

func (r *Repo) DailyRewardsOfDay(day civil.Date) (items []DayReward, err error) {
	err = r.db.Select(&items, `select * from day_reward where day = $1`, day)
	return
}
