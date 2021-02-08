package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dabankio/bbrpc"
	"github.com/dabankio/gobbc"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	perBatchBlocks = 20 //10000 default
	safeConfirms   = 60 //安全确认数，达到后认为区块不会回滚
)

// create user bbcsync IDENTIFIED BY 'bbc';
// CREATE DATABASE IF NOT EXISTS bbcsync default character set utf8mb4 collate utf8mb4_general_ci;
// grant ALL PRIVILEGES on bbcsync.* to bbcsync;
//
// delete from Block;
// delete from Tx;
func main() {
	conf := parseConf()
	db, err := newDB(conf.DB)
	pe(err)
	defer db.Close()

	client, err := bbrpc.NewClientWith(&bbrpc.ConnConfig{
		Host:       conf.RPCUrl,
		User:       conf.RPCUsr,
		Pass:       conf.RPCPassword,
		DisableTLS: true,
	}, &http.Client{
		Timeout:   time.Second * 15,
		Transport: &http.Transport{MaxIdleConnsPerHost: 1},
	})
	pe(err)
	defer client.Shutdown()

	for {
		var syncToHeight int //需要同步到的高度, 如果同步高度与最新高度很大则从数据库高度+n， 否则处理到最新高度
		rpcHeight, err := client.Getforkheight(nil)
		pe(err)
		dbTopHeightBlc, err := getMaxHeightBlock(db)
		dbHeight := int(dbTopHeightBlc.Height)
		if err != nil {
			if err == sql.ErrNoRows { //创世高度
				syncToHeight = 1
				err = nil
			} else {
				panic(err)
			}
		} else {
			if rpcHeight < dbHeight {
				panic(errors.Errorf("节点高度%d低于数据库最高记录%d", rpcHeight, dbHeight))
			} else if rpcHeight == dbHeight { //已经同步到最新高度了
				syncToHeight = -1
			} else {
				if rpcHeight-dbHeight > perBatchBlocks { //高度差过大则扫n个块
					syncToHeight = dbHeight + perBatchBlocks
				} else {
					syncToHeight = rpcHeight
				}
			}
		}

		if syncToHeight > 0 {
			syncToHash, err := client.Getblockhash(syncToHeight, nil)
			pe(err)
			safeHeight := rpcHeight - safeConfirms
			lowerHeight := dbHeight - 2   //随便减点，比dbHeight小就行
			if lowerHeight > safeHeight { //下限不会高于安全高度
				lowerHeight = safeHeight
			}
			if lowerHeight < 0 {
				lowerHeight = 0
			}
			log.Printf("exec task %d (lower:%d)\n", syncToHeight, lowerHeight)
			err = execTask(db, client, syncToHash[0], lowerHeight)
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
		//不知道为什么 用 ? 不行
		if *useful {
			sql += fmt.Sprintf(" and is_useful = 1")
		} else {
			sql += fmt.Sprintf(" and is_useful = 0")
		}
		// args = append(args, *useful)
	}
	err := db.Get(&ret, sql, args...)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func getVote(tplHex string) (dele, voter string) {
	if len(tplHex) != 132 {
		panic("not vote info len 132")
	}
	del, err := gobbc.NewCDestinationFromHexString(tplHex[:66])
	pe(err)
	owner, err := gobbc.NewCDestinationFromHexString(tplHex[66:])
	pe(err)
	return del.String(), owner.String()
}

func getTx(db *sqlx.Tx, sql string, args []interface{}) (tx Tx, err error) {
	err = db.Get(&tx, sql, args...)
	return
}

const voteAddrPrefix = "20w0"

// 标记交易的输入tx的spend_txid
// 入库tx(每个vout会产生一条记录)(如果是投票则处理相关字段)
func insertTx(dbTx *sqlx.Tx, blockHash string, tx *bbrpc.NoneSerializedTransaction) error {
	vinAmount := decimal.NewFromInt(0)
	for _, in := range tx.Vin {
		inTx, err := getTx(dbTx, "select id,amount,`to` from Tx where txid = ? and n = ?", []interface{}{in.Txid, in.Vout})
		if err != nil {
			return errors.Wrapf(err, "get tx:%s", in.Txid)
		}
		vinAmount = vinAmount.Add(inTx.Amount)
		_, err = dbTx.Exec("update Tx set spend_txid = ? where id = ?", tx.Txid, inTx.ID)
		if err != nil {
			return errors.Wrap(err, "update spend tx")
		}
	}
	var dposIn, clientIn, dposOut, clientOut string
	if strings.HasPrefix(tx.Sendto, voteAddrPrefix) { //投票
		dposIn, clientIn = getVote(tx.Sig[:132])
	}
	if strings.HasPrefix(tx.Sendfrom, voteAddrPrefix) {
		tplData := tx.Sig[0:132]
		if strings.HasPrefix(tx.Sendto, voteAddrPrefix) { //转投其他
			tplData = tx.Sig[132:264]
		}
		dposOut, clientOut = getVote(tplData)
	}
	data := tx.Data
	if len(data) >= 4096 {
		data = data[:4096]
	}
	sql := "insert Tx(block_hash,txid,form,`to`,amount,free,type,lock_until,n,data,dpos_in,client_in,dpos_out,client_out)values(?,?,?,?,?,?,?,?,0,?,?,?,?,?)"
	// [block_id,tx["txid"], tx["sendfrom"],tx["sendto"],tx["amount"],tx["txfee"],tx["type"],tx["lockuntil"],data,dpos_in,client_in,dpos_out,client_out]
	_, err := dbTx.Exec(sql, blockHash, tx.Txid, tx.Sendfrom, tx.Sendto, tx.Amount, tx.Txfee, tx.Type, tx.Lockuntil, data, dposIn, clientIn, dposOut, clientOut)
	if err != nil {
		return errors.Wrap(err, "insert tx")
	}
	amountFee := decimal.NewFromFloat(tx.Amount).Add(decimal.NewFromFloat(tx.Txfee))
	if amountFee.LessThan(vinAmount) {
		amount := vinAmount.Sub(amountFee)
		sql = "insert Tx(block_hash,txid,form,`to`,amount,free,type,lock_until,n,data)values(?,?,?,?,?,?,?,?,1,?)"
		_, err := dbTx.Exec(sql, blockHash, tx.Txid, tx.Sendfrom, tx.Sendfrom, amount, 0, tx.Type, 0, data)
		if err != nil {
			return errors.Wrap(err, "insert change tx")
		}
	}
	return err
}

//回滚单个块，标记无效，处理tx
func rollBackBlock(db *sqlx.DB, blockHash string) error {
	log.Println("rollback block:", blockHash)
	err := runInTx(db, func(tx *sqlx.Tx) error {
		var err error
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
			_, err = tx.Exec(fmt.Sprintf("update Tx set spend_txid = null where spend_txid = '%s'", t.Txid))
			if err != nil {
				return err
			}
			_, err = tx.Exec(fmt.Sprintf("Delete from Tx where txid = '%s'", t.Txid))
			if err != nil {
				return err
			}
		}
		return err
	})
	return err
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

// 如果区块在库里存在则标记为useful, 否则入库block, 入库交易
func useful(db *sqlx.DB, client *bbrpc.Client, blockHash string) error {
	det, err := client.Getblockdetail(blockHash)
	if err != nil {
		return errors.Wrap(err, "rpc get block detail")
	}
	log.Println("useful block", det.Height, blockHash)
	_, err = getBlock(db, blockHash, nil)

	err = runInTx(db, func(dbTx *sqlx.Tx) error {
		//先处理block
		if err != nil {
			if err != sql.ErrNoRows {
				return errors.Wrap(err, "get block detail")
			}
			//这个块在数据库里没有
			sql := "insert into Block(hash,prev_hash,time,height,reward_address,bits,reward_money,type,fork_hash) values(?,?,?,?,?,?,?,?,?)"
			_, err = dbTx.Exec(sql, det.Hash, det.HashPrev, det.Time, det.Height, det.Txmint.Sendto, det.Bits, det.Txmint.Amount, det.Type, det.Fork)
			if err != nil {
				return errors.Wrap(err, "insert block")
			}
		} else { //这个块在数据库里已经存在了
			_, err = dbTx.Exec(fmt.Sprintf("update Block set is_useful = 1 where `hash` = '%s'", blockHash))
			if err != nil {
				return errors.Wrap(err, "update block")
			}
		}
		// 再处理tx
		for _, tx := range append(det.Tx, det.Txmint) {
			err = insertTx(dbTx, blockHash, &tx)
			if err != nil {
				return errors.Wrap(err, "insert tx")
			}
		}
		return err
	})
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

//根据高块和【当前找到的最高有效块】的高度，回滚或标记块有效
//usefulBlc 【当前找到的最高有效块】
//如果有效块低于高块， 高块到有效块之间的块要标记失效
func updateState(db *sqlx.DB, client *bbrpc.Client, usefulBlc *Block) error {
	log.Println("update state:", usefulBlc.Height, usefulBlc.Hash)
	prevHash, height := usefulBlc.Hash, usefulBlc.Height
	p3hash, _ := prevHash, height

	rollBackHash, useBlockHash := []string{}, []string{}
	endBlock, err := getMaxHeightBlock(db)
	if err != nil {
		return err
	}
	p2hash, p2height := endBlock.Hash, endBlock.Height
	if endBlock.Hash == p3hash { //这个块就是库中的最高块
		return nil
	}
	if endBlock.Height > usefulBlc.Height { //库中最高块高于【当前找到的最高有效块】，则库中新块(【当前找到的最高有效块】到库中的最高块)要回滚
		rollBackHash = append(rollBackHash, p2hash)
		for {
			prevBlock, err := getPrevBlock(db, p2hash)
			if err != nil {
				return err
			}
			p2hash, p2height = prevBlock.Hash, prevBlock.Height
			if p2height == usefulBlc.Height {
				break
			}
			rollBackHash = append(rollBackHash, p2hash)
		}
	} else if usefulBlc.Height > endBlock.Height { //【当前找到的最高有效块】高于数据库高块，则直到高块都为有效块
		prevHeight := usefulBlc.Height
		useBlockHash = append(useBlockHash, p3hash)
		for {
			prevBlock, err := getPrevBlock(db, p3hash)
			if err != nil {
				return err
			}
			prevHeight = prevBlock.Height
			if prevHeight == endBlock.Height {
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

// lowerHeight 比数据库最新高度低，比安全高度低
func execTask(db *sqlx.DB, client *bbrpc.Client, sync2blockHash string, lowerHeight int) error {
	blockHashsNeed2sync := []string{} //需要同步的区块hash
	var dbBlc *Block
	dbBlc, err := getBlock(db, sync2blockHash, &usefulBlockBool) //如果数据库没有同步到这个block,那么此时dbBlc是nil
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	for { //loop 从sync2blockHash往前找，直到 创世块或lowerHeight或者在数据库找到了有效的块， 中间的blockHash 为需要同步的区块
		blockHashsNeed2sync = append(blockHashsNeed2sync, sync2blockHash)
		rpcBlk, err := client.Getblock(sync2blockHash)
		if err != nil {
			return err
		}
		if rpcBlk.Height == 1 || rpcBlk.Height <= uint(lowerHeight) { //创世块, 或者达到安全高度
			break
		}
		sync2blockHash = rpcBlk.Prev
		dbBlc, err = getBlock(db, sync2blockHash, &usefulBlockBool)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			dbBlc = nil
		} else { //找到了有效的块（dbBlc 非 nil）
			break
		}
	}

	if dbBlc != nil { //在数据库找到了有效的块
		if err = updateState(db, client, dbBlc); err != nil {
			return err
		}
	}
	for i, j := 0, len(blockHashsNeed2sync)-1; i < j; i, j = i+1, j-1 { //reverse
		blockHashsNeed2sync[i], blockHashsNeed2sync[j] = blockHashsNeed2sync[j], blockHashsNeed2sync[i]
	}
	for _, useHash := range blockHashsNeed2sync {
		if err = useful(db, client, useHash); err != nil {
			return errors.Wrap(err, "useful err")
		}
	}
	return nil
}

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
