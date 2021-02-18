package sync

import (
	"bbcsyncer/infra"
	"context"
	"database/sql"
	"log"

	"github.com/dabankio/bbrpc"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewWorker(repo *Repo, client *bbrpc.Client) *Worker {
	return &Worker{
		repo: repo, client: client,
	}
}

// Worker sync worker,处理同步过程
// 工作原理：
// 扫链至最新的块，区块数、交易等数据写入数据库
// 每隔1分钟扫描一次，扫描开始前移除分叉的块的数据
type Worker struct {
	repo   *Repo
	client *bbrpc.Client
}

func (w *Worker) Sync(ctx context.Context) {
	// log.Println("worker sync")
	// defer log.Println("sync done")

	err := w.removeForkedBlocks()
	if err != nil {
		log.Println("[err] ", err)
		return
	}
	err = w.sync2latest(ctx)
	if err != nil {
		log.Println("[err] ", err)
		return
	}
}

// 移除分叉的块
func (w *Worker) removeForkedBlocks() error {
	// log.Println("will removed discarded blocks")
	var removedBlocks []uint64
	defer func() {
		if l := len(removedBlocks); l > 0 {
			log.Printf("%d blocks-removed, %d~%d", l, removedBlocks[0], removedBlocks[l-1])
		}
	}()
	for {
		checkBlock, err := w.repo.LastestSyncedBlock()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		block, err := w.client.GetblockByHeight(checkBlock.Height, nil)
		if err != nil {
			return errors.Wrap(err, "get block by height err")
		}
		if block.Hash != checkBlock.Hash {
			log.Println("block_deleted, height: ", checkBlock.Height)
			err = infra.RunInTx(w.repo.db, func(tx *sqlx.Tx) error {
				return w.repo.deleteBlock(tx, checkBlock.Height)
			})
			if err != nil {
				return errors.Wrapf(err, "delete block(%d) err", checkBlock.Height)
			}
			removedBlocks = append(removedBlocks, checkBlock.Height)
		} else { //高度上的区块hash和rpc一致，认为截止这个区块是没有分叉的
			return nil
		}
	}
}

// 同步至最新高度
func (w *Worker) sync2latest(ctx context.Context) error {
	var nextBlockHeight uint64
	lastBlock, err := w.repo.LastestSyncedBlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			nextBlockHeight = 0
		} else {
			return err
		}
	} else {
		nextBlockHeight = lastBlock.Height + 1
	}

	topHeight, err := w.client.Getforkheight(nil)
	if err != nil {
		return errors.Wrap(err, "get top height err")
	}
	if uint64(topHeight) < nextBlockHeight { //没有新的块
		log.Println("no new block")
		return nil
	}
	// log.Printf("will sync, (%d -> %d]\n", nextBlockHeight-1, topHeight)

	type detailOrErr struct {
		detail *bbrpc.BlockDetail //为空时表示结束
		err    error
	}

	//新开goroutine不停的查询区块数据，主goroutine不停的写入数据以提高效率
	detailChan := make(chan detailOrErr, 100)
	defer close(detailChan)

	rpcCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func(_ctx context.Context) {
		fnSend := func(x detailOrErr) {
			select {
			case <-_ctx.Done():
				return
			default: //do nothing
				detailChan <- x
			}
		}
		for ; nextBlockHeight <= uint64(topHeight); nextBlockHeight++ {
			select {
			case <-_ctx.Done():
				return
			default: //do nothing
			}
			hash, err := w.client.Getblockhash(int(nextBlockHeight), nil)
			if err != nil {
				fnSend(detailOrErr{err: errors.Wrap(err, "getblock hash err")})
				return
			}
			if len(hash) == 0 {
				fnSend(detailOrErr{err: errors.Errorf("no block hash %d", nextBlockHeight)})
				return
			}
			detail, err := w.client.Getblockdetail(hash[0])
			if err != nil {
				fnSend(detailOrErr{err: errors.Wrap(err, "get block detail err")})
				return
			}
			fnSend(detailOrErr{detail: detail})
		}
		fnSend(detailOrErr{detail: nil})
	}(rpcCtx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case detailErr := <-detailChan:
			if detailErr.err != nil {
				return err
			}
			if detailErr.detail == nil {
				return nil
			}
			err = w.saveBlock(detailErr.detail)
			if err != nil {
				return errors.Wrap(err, "save block err")
			}
		}
	}
}

func (w *Worker) saveBlock(bd *bbrpc.BlockDetail) error {
	log.Printf("save_block %7d %s, tx count: %d", bd.Height, bd.Hash, len(bd.Tx))
	return infra.RunInTx(w.repo.db, func(tx *sqlx.Tx) error {
		err := w.repo.insertBlock(tx, NewBlock(bd))
		if err != nil {
			return errors.Wrap(err, "insert block err")
		}
		for _, t := range NewTxsFromBlock(bd) {
			if err = w.repo.insertTx(tx, t); err != nil {
				return errors.Wrap(err, "insert tx err")
			}
		}
		return err
	})
}
