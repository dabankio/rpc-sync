package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/dabankio/bbrpc"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"´

	_ "github.com/go-sql-driver/mysql"
)

const (
	perBatchBlocks = 20 //10000 default
	safeConfirms   = 60 //安全确认数，达到后认为区块不会回滚
)

// create user bbcsync IDENTIFIED BY 'bbc';
// CREATE DATABASE IF NOT EXISTS bbcsync default character set utf8mb4 collate utf8mb4_general_ci;
// grant ALL PRIVILEGES on bbcsync.* to bbcsync;
//
// delete * from Block;
// delete from Tx;
func main() {
	conf := parseConf()
	db, err := newDB(conf.DB)
	pe(err)
	defer db.Close()

	client, err := bbrpc.NewClient(&bbrpc.ConnConfig{
		Host:       conf.RPCUrl,
		User:       conf.RPCUsr,
		Pass:       conf.RPCPassword,
		DisableTLS: true,
	})
	pe(err)
	defer client.Shutdown()

	for {
		var execHeight int
		rpcHeight, err := client.Getforkheight(nil)
		pe(err)
		dbTopHeightBlk, err := getMaxHeightBlock(db)
		dbHeight := int(dbTopHeightBlk.Height)
		if err != nil {
			if err == sql.ErrNoRows { //创世高度
				execHeight = 1
			} else {
				panic(err)
			}
		} else {
			if rpcHeight < dbHeight {
				panic(errors.Errorf("节点高度%d低于数据库最高记录%d", rpcHeight, dbHeight))
			} else if rpcHeight == dbHeight {
				execHeight = -1
			} else {
				if rpcHeight-dbHeight > perBatchBlocks { //高度差过大则扫n个块
					execHeight = dbHeight + perBatchBlocks
				} else {
					execHeight = rpcHeight
				}
			}
		}
		pe(err)

		// h, err := getForkHeight(db, client)
		pe(err)
		if execHeight > 0 {
			blockHash, err := client.Getblockhash(execHeight, nil)
			pe(err)
			safeHeight := rpcHeight - safeConfirms
			lowerHeight := dbHeight - 2   //随便减点，比dbHeight小就行
			if lowerHeight > safeHeight { //下限不会高于安全高度
				lowerHeight = safeHeight
			}
			log.Printf("exec task %d (lower:%d)", execHeight, lowerHeight)
			err = execTask(db, client, blockHash[0], lowerHeight)
			pe(err)
		} else {
			log.Println("sleeping")
			time.Sleep(3 * time.Second)
		}
	}
}

// Conf .
type Conf struct {
	DB          string `json:"db,omitempty"`
	RPCUrl      string `json:"rpc_url,omitempty"`
	RPCUsr      string `json:"rpc_usr"`
	RPCPassword string `json:"rpc_password,omitempty"`
}

type Block struct {
	ID           int64 `db:"id"`
	Time, Height int64
	IsUseful     types.BitBool
	Bits         int
	RewardState  types.BitBool

	Hash, ForkHash, PrevHash, Type, RewardAddress, RewardMoney string
}

type Tx struct {
	ID           int64 `db:"id"`
	LockUntil    int64
	N            int             //vout index
	Amount, Free decimal.Decimal //free: fee ...

	BlockHash, Txid, From, To, Type, SpendTxid, Data, DposIn, ClientIn, DposOut, ClientOut string
}

var confFile string

func parseConf() (c Conf) {
	if !flag.Parsed() {
		flag.StringVar(&confFile, "conf", "./dev.env.json", "-conf=/etc/sync_conf.json")
		flag.Parse()
	}
	b, err := ioutil.ReadFile(confFile)
	pe(err)
	pe(json.Unmarshal(b, &c))
	return
}

func pe(err error) {
	if err != nil {
		panic(err)
	}
}

func newDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.MapperFunc(func(s string) string { //snake_case
		var ret string
		for i, r := range s {
			if r >= 'A' && r <= 'Z' && i > 0 {
				ret += "_"
			}
			ret += strings.ToLower(string(r))
		}
		return ret
	})
	return db, nil
}

// db block by hash
func getBlock(db *sqlx.DB, blockHash string, useful *types.BitBool) (*Block, error) {
	var ret Block
	args := []interface{}{blockHash}
	sql := "select * from Block where hash = ?"
	if useful != nil {
		sql += fmt.Sprintf(" and is_useful = ?")
		args = append(args, *useful)
	}
	err := db.Get(&ret, sql, args...)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func getVote(tplHex string) (string, string) {
	// TODO
	return "not_impl: dposAddr", "client_addr"
}

func getTx(db *sqlx.DB, sql string, args []interface{}) (tx Tx, err error) {
	err = db.Get(&tx, sql, args...)
	return
}

const voteAddrPrefix = "20w0"

// 标记交易的输入tx的spend_txid,入库tx(每个vout会产生一条记录)
func insertTx(db *sqlx.DB, blockHash string, tx *bbrpc.NoneSerializedTransaction) error {
	vinAmount := decimal.NewFromInt(0)
	for _, in := range tx.Vin {
		inTx, err := getTx(db, "select id,amount,`to` from Tx where txid = ? and n = ?", []interface{}{in.Txid, in.Vout})
		if err != nil {
			return errors.Wrapf(err, "get tx:%s", in.Txid)
		}
		vinAmount = vinAmount.Add(inTx.Amount)
		_, err = db.Exec("update Tx set spend_txid = ? where id = ?", tx.Txid, inTx.ID)
		if err != nil {
			return errors.Wrap(err, "update spend tx")
		}
	}
	var dposIn, clientIn, dposOut, clientOut string
	if strings.HasPrefix(tx.Sendto, voteAddrPrefix) { //投票
		dposIn, clientIn = getVote(tx.Sig[:132])
	}
	if strings.HasPrefix(tx.Sendfrom, voteAddrPrefix) {
		if strings.HasPrefix(tx.Sendto, voteAddrPrefix) { //转投其他
			dposOut, clientOut = getVote(tx.Sig[132:264])
		} else {
			dposOut, clientOut = getVote(tx.Sig[0:132])
		}
	}
	data := tx.Data
	if len(data) >= 4096 {
		data = data[:4096]
	}
	sql := "insert Tx(block_hash,txid,form,`to`,amount,free,type,lock_until,n,data,dpos_in,client_in,dpos_out,client_out)values(?,?,?,?,?,?,?,?,0,?,?,?,?,?)"
	// [block_id,tx["txid"], tx["sendfrom"],tx["sendto"],tx["amount"],tx["txfee"],tx["type"],tx["lockuntil"],data,dpos_in,client_in,dpos_out,client_out]
	_, err := db.Exec(sql, blockHash, tx.Txid, tx.Sendfrom, tx.Sendto, tx.Amount, tx.Txfee, tx.Type, tx.Lockuntil, data, dposIn, clientIn, dposOut, clientOut)
	if err != nil {
		return errors.Wrap(err, "insert tx")
	}
	amountFee := decimal.NewFromFloat(tx.Amount).Add(decimal.NewFromFloat(tx.Txfee))
	if amountFee.LessThan(vinAmount) {
		amount := vinAmount.Sub(amountFee)
		sql = "insert Tx(block_hash,txid,form,`to`,amount,free,type,lock_until,n,data)values(?,?,?,?,?,?,?,?,1,?)"
		_, err := db.Exec(sql, blockHash, tx.Txid, tx.Sendfrom, tx.Sendfrom, amount, 0, tx.Type, 0, data)
		if err != nil {
			return errors.Wrap(err, "insert change tx")
		}
	}
	return err
}

//回滚单个块，标记无效，处理tx
func rollBackBlock(db *sqlx.DB, blockHash string) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		}
	}()

	_, err = tx.Exec(fmt.Sprintf("update Block set is_useful = 0 where `hash` = '%s'", blockHash))
	if err != nil {
		return err
	}
	var txs []Tx
	err = tx.Select(&txs, fmt.Sprintf("SELECT txid from Tx where block_hash = '%s' ORDER BY id desc", blockHash))
	if err != nil {
		return err
	}
	for _, t := range txs {
		_, err = db.Exec(fmt.Sprintf("update Tx set spend_txid = null where spend_txid = '%s'", t.Txid))
		if err != nil {
			return err
		}
		_, err = db.Exec(fmt.Sprintf("Delete from Tx where txid = '%s'", t.Txid))
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

// 如果区块在库里存在则标记为useful, 否则入库block, 入库交易
func useful(db *sqlx.DB, client *bbrpc.Client, blockHash string) error {
	det, err := client.Getblockdetail(blockHash)
	if err != nil {
		return errors.Wrap(err, "rpc get block detail")
	}
	log.Println("useful block", det.Height, blockHash)
	_, err = getBlock(db, blockHash, nil)
	if err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "get block detail")
		}
		sql := "insert into Block(hash,prev_hash,time,height,reward_address,bits,reward_money,type,fork_hash) values(?,?,?,?,?,?,?,?,?)"
		_, err = db.Exec(sql, det.Hash, det.HashPrev, det.Time, det.Height, det.Txmint.Sendto, det.Bits, det.Txmint.Amount, det.Type, det.Fork)
		if err != nil {
			return errors.Wrap(err, "insert block")
		}
	} else {
		_, err = db.Exec(fmt.Sprintf("update Block set is_useful = 1 where `hash` = '%s'", blockHash))
		if err != nil {
			return errors.Wrap(err, "update block")
		}
	}
	for _, tx := range det.Tx {
		err = insertTx(db, blockHash, &tx)
		if err != nil {
			return errors.Wrap(err, "insert tx")
		}
	}
	err = insertTx(db, blockHash, &det.Txmint)
	if err != nil {
		return errors.Wrap(err, "insert tx")
	}
	check()
	return nil
}

// 数据库中的max(height) block
func getMaxHeightBlock(db *sqlx.DB) (blk Block, err error) {
	err = db.Get(&blk, "SELECT `hash`, prev_hash,height from Block ORDER BY id DESC LIMIT 1")
	return
}

// 数据库中 blockHash 的前一个块
func getPrevBlock(db *sqlx.DB, blockHash string) (blk Block, err error) {
	err = db.Get(&blk, `SELECT b2.hash, b2.prev_hash, b2.height 
	from Block b1 
	inner JOIN Block b2 on b1.prev_hash = b2.hash 
	where b1.hash = ?`, blockHash)
	return
}

//根据高块和入参的高度，回滚或标记块有效
func updateState(db *sqlx.DB, client *bbrpc.Client, blc *Block) error {
	prevHash, height := blc.Hash, blc.Height
	p3hash, p3height := prevHash, height

	rollBackHash, useBlockHash := []string{}, []string{}
	endBlock, err := getMaxHeightBlock(db)
	if err != nil {
		return err
	}
	p2hash, p2height := endBlock.Hash, endBlock.Height
	if p2hash == p3hash { //这个块就是库中的最高块
		return nil
	}
	if p2height > p3height { //库中最高块高于入参块，则库中新块要回滚（直到入参height）
		rollBackHash = append(rollBackHash, p2hash)
		for {
			prevBlock, err := getPrevBlock(db, p2hash)
			if err != nil {
				return err
			}
			p2hash, p2height = prevBlock.Hash, prevBlock.Height
			if p2height == p3height {
				break
			}
			rollBackHash = append(rollBackHash, p2hash)
		}
	} else if p3height > p2height { //入参高于数据库高块，则直到高块都为有效块
		useBlockHash = append(useBlockHash, p3hash)
		for {
			prevBlock, err := getPrevBlock(db, p3hash)
			if err != nil {
				return err
			}
			p3height = prevBlock.Height
			if p3height == p2height {
				break
			}
			p3hash = prevBlock.Hash
			useBlockHash = append(useBlockHash, p3hash)
		}
	}

	for p2hash != p3hash {
		rollBackHash = append(rollBackHash, p2hash)
		prevBlk2, err := getPrevBlock(db, p2hash)
		if err != nil {
			return err
		}
		prevBlk3, err := getPrevBlock(db, p3hash)
		if err != nil {
			return err
		}
		p2hash, p3hash = prevBlk2.Hash, prevBlk3.Hash
		if p2hash != p3hash {
			useBlockHash = append(useBlockHash, p2hash)
		}
	}

	for _, cancelHash := range rollBackHash {
		err = rollBackBlock(db, cancelHash)
		if err != nil {
			return err
		}
	}
	for i, j := 0, len(useBlockHash)-1; i < j; i, j = i+1, j-1 { //reverse slice
		useBlockHash[i], useBlockHash[j] = useBlockHash[j], useBlockHash[i]
	}
	for _, hash := range useBlockHash {
		err = useful(db, client, hash)
		if err != nil {
			return err
		}
	}
	return nil
}

var usefulBlockBool = types.BitBool(true)

func execTask(db *sqlx.DB, client *bbrpc.Client, blockHash string, lowerHeight int) error {
	taskAddHash := []string{}
	var dbBlc *Block
	dbBlc, err := getBlock(db, blockHash, &usefulBlockBool)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	for dbBlc == nil { //数据库没有这个区块,则尝试将blk绑定为前一个块,直到创世块
		taskAddHash = append(taskAddHash, blockHash)
		rpcBlk, err := client.Getblock(blockHash)
		if err != nil {
			return err
		}
		if rpcBlk.Height == 1 || rpcBlk.Height <= uint(lowerHeight) { //创世块
			break
		}
		blockHash = rpcBlk.Prev
		dbBlc, err = getBlock(db, blockHash, &usefulBlockBool)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			dbBlc = nil
		}
	}
	if dbBlc != nil {
		if err = updateState(db, client, dbBlc); err != nil {
			return err
		}
	}
	for i, j := 0, len(taskAddHash)-1; i < j; i, j = i+1, j-1 { //reverse
		taskAddHash[i], taskAddHash[j] = taskAddHash[j], taskAddHash[i]
	}
	for _, useHash := range taskAddHash {
		if err = useful(db, client, useHash); err != nil {
			return errors.Wrap(err, "useful err")
		}
	}
	return nil
}

// 当前节点高度（如果数据库差的多则数据库高度+10000 ？？
// func getForkHeight(db *sqlx.DB, client *bbrpc.Client) (int64, error) {
// 	rpcHeight, err := client.Getforkheight(nil)
// 	if err != nil {
// 		return 0, err
// 	}
// 	dbTopHeightBlk, err := getMaxHeightBlock(db)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return 1, nil
// 		}
// 		return 0, err
// 	}
// 	if int64(rpcHeight) > dbTopHeightBlk.Height {
// 		if x := dbTopHeightBlk.Height + perBatchBlocks; x < int64(rpcHeight) { //10000
// 			return x, nil
// 		}
// 		return int64(rpcHeight), nil
// 	}
// 	return 0, nil
// }

func getPool(db *sqlx.DB) (ret []string, err error) {
	err = db.Select(&ret, "select address from pool")
	return
}

func getListDelegate(db *sqlx.DB, client *bbrpc.Client) error {
	forks, err := client.Listfork(true)
	if err != nil {
		return err
	}
	dbForks, err := getPool(db)
	if err != nil {
		return err
	}
	for _, f := range forks {
		if strings.Contains(strings.Join(dbForks, ","), f.Fork) {
			continue
		}
		_, err = db.Exec(
			"insert pool(address,name,type,`key`,fee)values(?,?,?,?,?)",
			f.Fork, "", "dpos", "123456", 0.05,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func check() {
	if 2 > 1 {
		return
	}
}
