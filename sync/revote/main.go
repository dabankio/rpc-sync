package main

import (
	"bbcsyncer/sync"
	"log"
	"os"
)

//之前的投票计算算法有问题，重新算投票数据
// delete from dpos_vote;
func main() {
	const height_per_query = 100
	cursor := 0

	var maxHeight = 636519 //select max(height) + 100 from blocks;
	db, err := sync.NewPGDB(sync.Conf{
		DB: os.Getenv("DEV_DB"),
	})
	sync.PanicErr(err)

	for ; cursor < maxHeight; cursor += height_per_query {
		log.Printf("%d~%d\n", cursor, cursor+height_per_query)

		var txs []sync.Tx
		err := db.Select(&txs, `select block_height, txid, amount, txfee, typ, send_from, send_to, sig from txs where block_height >= $1 and block_height < $2`, cursor, cursor+height_per_query)
		sync.PanicErr(err)
		log.Println("insert votes")

		for _, tx := range txs {
			votes := tx.DposVotes()
			for _, v := range votes {
				_, err = db.Exec(`insert into dpos_vote (block_height, txid, delegate, voter, amount) values ($1, $2, $3, $4, $5)`,
					v.BlockHeight, v.Txid, v.Delegate, v.Voter, v.Amount)
				sync.PanicErr(err)
			}
		}
	}
}
