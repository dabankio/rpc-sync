package sync

import (
	"bbcsyncer/infra"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

type Repo struct {
	db *sqlx.DB
}

func (r *Repo) LastestSyncedBlock() (Block, error) {
	var blc Block
	err := r.db.Get(&blc, "select * from blocks order by height desc limit 1")
	return blc, infra.WrapErr(err, "db get latest synced block err")
}

func (r *Repo) BlockByHeight(height uint64) (Block, error) {
	var blc Block
	err := r.db.Get(&blc, "select * from blocks where height = $1", height)
	return blc, infra.WrapErr(err, "db get block by height err")
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

// VotesAtHeight 某一区块的投票数据
func (r *Repo) VotesAtHeight(height int) (items []DposVote, err error) {
	err = r.db.Select(&items, `select * from dpos_vote where block_height = $1`, height)
	return
}

// [from, to] (含from, to)高度区间内的区块
func (r *Repo) BlocksBetweenHeight(fromHeight, toHeight uint64) (items []Block, err error) {
	err = r.db.Select(&items, `
	select * from blocks 
	where height >= $1 and height <= $2 
	order by height`,
		fromHeight, toHeight)
	return
}

func (r *Repo) BlocksInHeight(height ...uint64) (items []Block, err error) {
	if len(height) == 0 {
		return nil, nil
	}
	err = r.db.Select(&items, `select * from blocks where height in ($1)`, height)
	return
}

// 2个高度间(含端点)的投票详情
func (r *Repo) DposVotesBetweenHeight(fromHeight, toHeight uint64) (items []DposVote, err error) {
	if fromHeight > toHeight {
		return nil, errors.New("from height is greater than to")
	}
	err = r.db.Select(&items, `select * from dpos_vote where block_height >= $1 and block_height <= $2`, fromHeight, toHeight)
	return
}

func (r *Repo) TxsOfHeight(height uint64) (txs []Tx, err error) {
	err = infra.WrapErr(r.db.Select(&txs, `select * from txs where block_height = $1`, height), "get txs of height err")
	return
}

func (r *Repo) WalkBlocks(walker func(*Block, []Tx) error) error {
	const height_per_query = 100
	var maxHeight uint64
	{ //get max height of blocks
		err := r.db.Get(&maxHeight, "select max(height) from blocks")
		infra.PanicErr(err)
		maxHeight += height_per_query
	}

	log.Println("max height:", maxHeight)
	var cursor uint64
	for ; cursor <= maxHeight; cursor++ { //xxx 可以改成一个goroutine读，一个处理以提高效率； 另外也可以处理为批量查询（一个块一个块的遍历有点慢）
		b, err := r.BlockByHeight(cursor)
		if err != nil {
			return err
		}
		txs, err := r.TxsOfHeight(cursor)
		if err != nil {
			return err
		}
		err = walker(&b, txs)
		if err != nil {
			return err
		}
	}
	return nil
}
