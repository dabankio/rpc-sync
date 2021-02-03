package sync

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/dabankio/bbrpc"
	"github.com/dabankio/gobbc"
	"github.com/shopspring/decimal"
)

type TxType string

const (
	TxTypeToken TxType = "token"
	TxTypeStake TxType = "stake"
)

func NewTxsFromBlock(block *bbrpc.BlockDetail) []Tx {
	var txs []Tx
	for _, t := range append(block.Tx, block.Txmint) {
		tx := Tx{
			BlockHeight: uint64(block.Height),

			Txid:      t.Txid,
			Version:   uint16(t.Version),
			Typ:       TxType(t.Type),
			Time:      time.Unix(int64(t.Time), 0),
			Lockuntil: uint32(t.Lockuntil),
			Anchor:    t.Anchor,
			Blockhash: t.Blockhash,
			Sendfrom:  t.Sendfrom,
			Sendto:    t.Sendto,
			Amount:    decimal.NewFromFloat(t.Amount),
			Txfee:     decimal.NewFromFloat(t.Txfee),
			Data:      t.Data,
			Sig:       t.Sig,
			Fork:      t.Fork,
		}
		for _, in := range t.Vin {
			tx.Vin = append(tx.Vin, VinPoint{
				Txid: in.Txid,
				Vout: in.Vout,
			})
		}
		txs = append(txs, tx)
	}
	return txs
}

type Tx struct {
	BlockHeight uint64

	Txid      string
	Version   uint16
	Typ       TxType
	Time      time.Time
	Lockuntil uint32
	Anchor    string
	Blockhash string
	Sendfrom  string
	Sendto    string
	Amount    decimal.Decimal
	Txfee     decimal.Decimal
	Data      string
	Sig       string
	Fork      string

	Vin VinPoints
}

type VinPoints []VinPoint

// VinPoint .
type VinPoint struct {
	Txid string `json:"txid"`
	Vout uint   `json:"vout"`
}

var _ driver.Value = VinPoints{}

// Value implements the driver.Valuer interface for database serialization.
func (vps VinPoints) Value() (driver.Value, error) {
	b, err := json.Marshal(vps)
	return string(b), err
}

var _ sql.Scanner = (*VinPoints)(nil)

// Scan implements the sql.Scanner interface for database deSerialization.
func (vps *VinPoints) Scan(src interface{}) error {
	var x VinPoints
	err := TryUnmarshal(src, &x)
	*vps = x
	return err
}

const voteAddrPrefix = "20w0"

func isVoteAddress(add string) bool { return strings.HasPrefix(add, voteAddrPrefix) }

func getVote(tplHex string) (delegate, voter string) {
	if len(tplHex) != 132 {
		panic("not vote info len 132")
	}
	del, err := gobbc.NewCDestinationFromString(tplHex[:66])
	PanicErr(err)
	owner, err := gobbc.NewCDestinationFromString(tplHex[66:])
	PanicErr(err)
	return del.String(), owner.String()
}

func (tx Tx) DposVotes() (votes []DposVote) {
	if tx.Typ == TxTypeStake { //dpos出块
		votes = append(votes, DposVote{
			BlockHeight: tx.BlockHeight,
			Txid:        tx.Txid,
			Delegate:    tx.Sendto,
			Voter:       tx.Sendto,
			Amount:      tx.Amount,
		})
		return
	}

	fnAppendVote := func(tplData string, amount decimal.Decimal) {
		delegate, voter := getVote(tplData)
		votes = append(votes, DposVote{
			BlockHeight: tx.BlockHeight,
			Txid:        tx.Txid,
			Delegate:    delegate,
			Voter:       voter,
			Amount:      amount,
		})
	}
	fromAddIsVoteTemplate, toAddIsVoteTemplate := isVoteAddress(tx.Sendfrom), isVoteAddress(tx.Sendto)
	amount := tx.Amount
	if fromAddIsVoteTemplate && toAddIsVoteTemplate { //撤回同时投到另一个节点
		fnAppendVote(tx.Sig[132:264], amount.Neg()) //撤票
		fnAppendVote(tx.Sig[:132], amount)          //投票
	} else if fromAddIsVoteTemplate { //仅撤回投票
		fnAppendVote(tx.Sig[:132], amount.Neg()) //撤票
	} else if toAddIsVoteTemplate { //仅投票
		fnAppendVote(tx.Sig[:132], amount) //投票
	} else {
		return
	}
	return
}

type DposVote struct {
	// fromHeight, toHeight, txids, memo
	// 对于合并的投票不能丢失信息（高度 txid 金额），合并投票针对规律性大量交易，合并投票需要考虑统计需求（按日处理），合并统计需要可以支持恢复或拆分以支持统计需求
	BlockHeight uint64
	Txid        string
	Delegate    string          //节点地址
	Voter       string          //投票人
	Amount      decimal.Decimal //投票额, 负数表示撤回投票
}
