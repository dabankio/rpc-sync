package sync

import (
	"bbcsyncer/infra"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

type Repo struct {
	db *sqlx.DB
}

func (r *Repo) lastestSyncedBlock() (Block, error) {
	var blc Block
	err := r.db.Get(&blc, "select * from blocks order by height desc limit 1")
	return blc, infra.WrapErr(err, "db get latest synced block err")
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
		err = r.InsertVote(dbTx, v)
		if err != nil {
			return errors.Wrap(err, "insert vote err")
		}
	}
	return nil
}

func (r *Repo) InsertVote(dbTx *sqlx.Tx, v DposVote) (err error) {
	const sql = `insert into dpos_vote (block_height, txid, delegate, voter, amount) values ($1, $2, $3, $4, $5)`
	args := []interface{}{v.BlockHeight, v.Txid, v.Delegate, v.Voter, v.Amount}
	if dbTx != nil {
		_, err = dbTx.Exec(sql, args...)
	} else {
		_, err = r.db.Exec(sql, args...)
	}
	return
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
