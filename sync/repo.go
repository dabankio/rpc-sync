package sync

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var cachedFieldMap sync.Map

func setDBMapper(db *sqlx.DB) {
	db.MapperFunc(func(s string) string { //snake_case
		v, ok := cachedFieldMap.Load(s)
		if ok {
			return v.(string)
		}
		var ret string
		for i, r := range s {
			if r >= 'A' && r <= 'Z' && i > 0 {
				ret += "_"
			}
			ret += strings.ToLower(string(r))
		}
		cachedFieldMap.Store(s, ret)
		return ret
	})
}

func runInTx(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			_ = tx.Rollback()
			panic(err)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func NewMysqlDB(conf Conf) (*sqlx.DB, error) {
	return sqlx.Connect("mysql", conf.DB)
}

func NewPGDB(conf Conf) (*sqlx.DB, error) {
	log.Println("connecting db...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, "postgres", conf.DB)
	if err != nil {
		return nil, err
	}
	setDBMapper(db)
	return db, nil
}

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

type Repo struct {
	db *sqlx.DB
}

func (r *Repo) lastestSyncedBlock() (Block, error) {
	var blc Block
	err := r.db.Get(&blc, "select * from blocks order by height desc limit 1")
	return blc, WrapErr(err, "db get latest synced block err")
}

func (r *Repo) insertBlock(tx *sqlx.Tx, b Block) error {
	// log.Printf("%#v\n", b)
	_, err := tx.Exec(`insert into blocks(height,hash,prev_hash,version,typ,time,fork,coinbase,miner,tx_count) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		b.Height, b.Hash, b.PrevHash, b.Version, b.Typ, b.Time, b.Fork, b.Coinbase, b.Miner, b.TxCount)
	return err
}

func (r *Repo) insertTx(dbTx *sqlx.Tx, tx Tx) error {
	_, err := dbTx.Exec(`insert into txs (block_height, txid, "version", typ, time, lockuntil, anchor, block_hash,
		send_from, send_to, amount, txfee, data, sig, fork, vin) values 
		($1,$2,$3,$4,$5,$6,$7,$8,$9,$10, $11, $12, $13, $14, $15, $16)`,
		tx.BlockHeight, tx.Txid, tx.Version, tx.Typ, tx.Time, tx.Lockuntil, tx.Anchor, tx.BlockHash,
		tx.SendFrom, tx.SendTo, tx.Amount, tx.Txfee, tx.Data, tx.Sig, tx.Fork, tx.Vin,
	)
	if err != nil {
		return errors.Wrap(err, "insert tx err")
	}
	votes := tx.DposVotes()
	for _, v := range votes {
		_, err = dbTx.Exec(`insert into dpos_vote (block_height, txid, delegate, voter, amount) values ($1, $2, $3, $4, $5)`,
			v.BlockHeight, v.Txid, v.Delegate, v.Voter, v.Amount)
		if err != nil {
			return errors.Wrap(err, "insert vote err")
		}
	}
	return nil
}

//移除块数据：块、交易、投票数据
func (r *Repo) deleteBlock(tx *sqlx.Tx, height uint64) error {
	for _, sql := range []string{
		`delete from dpos_vote where block_height = $1`,
		`delete from txs where block_height = $1`,
		`delete from blocks where height = $1`} {
		_, err := tx.Exec(sql, height)
		if err != nil {
			return err
		}
	}
	return nil
}
