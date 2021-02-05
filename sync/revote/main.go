package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/sync"
	"log"
	"os"
	"time"
)

//删除投票数据，根据原始tx重新算投票数据
func main() {
	db, err := infra.NewPGDB(infra.Conf{
		DB: os.Getenv("DEV_DB"),
	})
	infra.PanicErr(err)
	repo := sync.NewRepo(db)

	{ //delete old logs
		log.Println("警告：10s后删除dpos_vote数据")
		time.Sleep(10 * time.Second)
		ret, err := db.Exec("delete from dpos_vote")
		infra.PanicErr(err)
		rowsDeleted, err := ret.RowsAffected()
		infra.PanicErr(err)
		log.Println("rows deleted: ", rowsDeleted)
	}

	const height_per_query = 100
	var maxHeight int
	{ //get max height of blocks
		err = db.Get(&maxHeight, "select max(height) from blocks")
		infra.PanicErr(err)
		maxHeight += height_per_query
	}

	log.Println("max height:", maxHeight)
	cursor := 0
	for ; cursor <= maxHeight; cursor += height_per_query {
		var txs []sync.Tx
		err := db.Select(&txs, `select block_height, txid, amount, txfee, typ, send_from, send_to, sig from txs 
		where block_height >= $1 and block_height < $2`, cursor, cursor+height_per_query)
		infra.PanicErr(err)

		voteCount := 0
		for _, tx := range txs {
			votes := tx.DposVotes()
			voteCount += len(votes)
			for _, v := range votes {
				err = repo.InsertVote(nil, v)
				infra.PanicErr(err)
			}
		}
		log.Printf("height %d~%d, tx count %d, votes: %d\n", cursor, cursor+height_per_query, len(txs), voteCount)
	}
}
