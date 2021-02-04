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
	TxTypeToken         TxType = "token"   //一般转账
	TxTypeStake         TxType = "stake"   //dpos
	TxTypeWork          TxType = "work"    //pow
	TxTypeGenesis       TxType = "genesis" //创世交易
	TxTypeCertification TxType = "certification"
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
			BlockHash: t.Blockhash,
			SendFrom:  t.Sendfrom,
			SendTo:    t.Sendto,
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
	BlockHash string
	SendFrom  string
	SendTo    string
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

const (
	voteAddrPrefix_20w0     = "20w0" //一般投票地址前缀
	delegateAddrPrefix_20m0 = "20m0" //超级节点地址前缀
)

func isVoteAddress(add string) bool { return strings.HasPrefix(add, voteAddrPrefix_20w0) }

func getVote(tplHex string) (delegate, voter string) {
	if len(tplHex) != 132 {
		panic("not vote info len 132")
	}
	del, err := gobbc.NewCDestinationFromHexString(tplHex[:66])
	PanicErr(err)
	owner, err := gobbc.NewCDestinationFromHexString(tplHex[66:])
	PanicErr(err)
	return del.String(), owner.String()
}

/**
计票原理：根据from/to地址判断计票策略, 根据签名获取投票人和节点地址

撤票量一般包括转账金额+手续费
超级节点出块奖励是一种特殊的节点自投(也可以用通用逻辑处理)

certification类型的交易不用管(dpos 内部机制)

xxxx 表示其他地址, +a自投, -b自撤, +c一般投票, -d一般撤票

from\to| 20m0  | 20w0  | xxxx
20m0   | -b\+a | -b\+c | -b
20w0   | -d\+a | -d\+c | -d
xxxx   |    +a |    +c |  \

bbcrpc_sync=# select txid from txs where send_from like '20m0%' and send_to like '20w0%';
                               txid
------------------------------------------------------------------
 5eec7ca43d5a5b3ce32dbc87db7ff3d3279b72801bcfe0858ead1107eaf86289
 5eec775a1854440c3cbe7854803c1c6093e8b3ea87df8bad8a7b28a51d38a720
 5eec7bc55f95cff54434dbe55294f5bf49af07a7c369435224115716785eaf95
 5eec7b1efde12d499ef013ea9917061e0297243508dee6a713688ecc0a8e96fd
(4 rows)

bbcrpc_sync=# select txid from txs where send_from like '20w0%' and send_to like '20m0%';
                               txid
------------------------------------------------------------------
 5eec703c78040ec1b74d197f241611691e62c53a3527464a82ffdc86f2e067e6
 5eec6ff294129f7ae41caf7a9d41844c52b6279d4ef1d6f71e5f18ef535bbfc0
 5eec70787b40c0097a28dbcbd0862c918bbbd08bb3cc01648c7ef3a2cc1454a4
(3 rows)

bbcrpc_sync=# select * from txs where txid='5eec7ca43d5a5b3ce32dbc87db7ff3d3279b72801bcfe0858ead1107eaf86289' limit 1;
-[ RECORD 1 ]+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
block_height | 301713
txid         | 5eec7ca43d5a5b3ce32dbc87db7ff3d3279b72801bcfe0858ead1107eaf86289
version      | 1
typ          | token
time         | 2020-06-19 08:51:48+00
lockuntil    | 0
anchor       | 00000000b0a9be545f022309e148894d1e1c853ccac3ef04cb6f5e5c70f41a70
block_hash   | 00049a91d12cfdaeabc605455998205346bdb7c1792b203fa088634d9856c004
send_from    | 20m0emvkr82b7qn8hq2fc8tt2135tph52ghnkqs8ggm06qfn42d6zxheq
send_to      | 20w01rscsdy4p3sdv1g2m1dvp8e2rgxvzze35s0tchcztc0bjsvmew6h5
amount       | 24300000
txfee        | 0.01
data         |
sig          | 020500ea6e7840967bd511b89ec46b4208cbab44a2846b3be51085006bbea4134d02020027734b41cb79f99187626ae79a9964f446d314e518e32dd61e018f75369832edc5cf1a21d68f1d14f674dee7a832849d88353614a04014169ce75cd34578020200ccead3c423d5d94691e20245183ca50f5c3ed99daceef6d837c5e038949602030000000000000041754267e847d98055065bfb64faabf18c1684ad29bf2e88a6f1dd63a4357123010edda02f36251344490c4b55ad454c29e1c00b2dd24234e8ef61e98a1f3343b001dc9c1428f3b559c14ec61504f21edc39777865aaa79e098508928e039f7f6dd1010351c2b5ec18ca5fbb1d8738cfb3aee6b1177f71a9c14a8001ebc9ea6d59f1ccb2884e21b921d89082c75f63fa63c99719291adca69ec432a647d0224b8a9be20bf6b1404ded8117785c9ddc24ca5b7dab9466315043c6db9e0a4d2c4324106d6ae7e9fa6afebc3edf65b9b9b68adc94c9f5be008f2086da8ccbda88fc4bc75b01
fork         | 00000000b0a9be545f022309e148894d1e1c853ccac3ef04cb6f5e5c70f41a70
vin          | [{"txid": "5ebe205b16c3ad8012991040134fcf42220c374d1ed8a7abaedef1988b788be6", "vout": 1}, {"txid": "5ece3a0023d79a405e2aa61af0396e99fa76c7a0843f6b801a2eb4b80c70fec2", "vout": 0}, {"txid": "5eec38a9aa1caa80567effeefec714fdf3a0effe565d6e52372c64e0a72341d9", "vout": 0}]

bbcrpc_sync=# select * from txs where txid='5eec703c78040ec1b74d197f241611691e62c53a3527464a82ffdc86f2e067e6';
-[ RECORD 1 ]+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
block_height | 301647
txid         | 5eec703c78040ec1b74d197f241611691e62c53a3527464a82ffdc86f2e067e6
version      | 1
typ          | token
time         | 2020-06-19 07:58:52+00
lockuntil    | 0
anchor       | 00000000b0a9be545f022309e148894d1e1c853ccac3ef04cb6f5e5c70f41a70
block_hash   | 00049a4fe29fc87f48a4ebf777754fa92e598131e9301ddb3b2e72de7870468d
send_from    | 20w01rscsdy4p3sdv1g2m1dvp8e2rgxvzze35s0tchcztc0bjsvmew6h5
send_to      | 20m0emvkr82b7qn8hq2fc8tt2135tph52ghnkqs8ggm06qfn42d6zxheq
amount       | 1
txfee        | 0.01
data         |
sig          | 020500ea6e7840967bd511b89ec46b4208cbab44a2846b3be51085006bbea4134d02020027734b41cb79f99187626ae79a9964f446d314e518e32dd61e018f753698020300000000000000651748ed9b930be5747851c486fc2fff79edb82cf38e785051cf28ee0be2a916013e18d4adbab6912214059ea8158bec4eb9d139ee2ea367781e3296b95783de4a0194f691f6d0cdde6078ce4fe2fcd6df73c1962c0b9c5355a515aa9a7b9b5842860105758eec69e4068380a48bbbbe87933f1627b8ce750f4e2eb53e65c50164a4c8f8cf01c14852e2034ba3bef2640d4daa599f98d3eaba1aba5183670639adf9390638229ab3854920c889ecab54f6347d248b2d01b9b95016dd70380d16e158c920a9d898962eb239cb1d0a83028441b4c975ea8295610a7b5cacd00775f826bc0a
fork         | 00000000b0a9be545f022309e148894d1e1c853ccac3ef04cb6f5e5c70f41a70
vin          | [{"txid": "5eec6b376a1a724a0462bbad4d52b7cc75d4fc84958bebd83515cca559002922", "vout": 0}]

bigbang> validateaddress 20m0emvkr82b7qn8hq2fc8tt2135tph52ghnkqs8ggm06qfn42d6zxheq
{
    "isvalid" : true,
    "addressdata" : {
        "address" : "20m0emvkr82b7qn8hq2fc8tt2135tph52ghnkqs8ggm06qfn42d6zxheq",
        "ismine" : true,
        "type" : "template",
        "template" : "delegate",
        "templatedata" : {
            "type" : "delegate",
            "hex" : "050032edc5cf1a21d68f1d14f674dee7a832849d88353614a04014169ce75cd34578020200ccead3c423d5d94691e20245183ca50f5c3ed99daceef6d837c5e0389496",
            "delegate" : {
                "delegate" : "7845d35ce79c161440a0143635889d8432a8e7de74f6141d8fd6211acfc5ed32",
                "owner" : "2080cstpkrghxbpa6j7h04h8r7jjgyq1yv6etsvqpv0vwbr1rjjb291cq"
            }
        }
    }
}
bigbang> validateaddress 20w01rscsdy4p3sdv1g2m1dvp8e2rgxvzze35s0tchcztc0bjsvmew6h5
{
    "isvalid" : true,
    "addressdata" : {
        "address" : "20w01rscsdy4p3sdv1g2m1dvp8e2rgxvzze35s0tchcztc0bjsvmew6h5",
        "ismine" : true,
        "type" : "template",
        "template" : "vote",
        "templatedata" : {
            "type" : "vote",
            "hex" : "0700020500ea6e7840967bd511b89ec46b4208cbab44a2846b3be51085006bbea4134d02020027734b41cb79f99187626ae79a9964f446d314e518e32dd61e018f753698",
            "vote" : {
                "delegate" : "20m0emvkr82b7qn8hq2fc8tt2135tph52ghnkqs8ggm06qfn42d6zxheq",
                "owner" : "20802ewtb875qkychgxh6nswtk5jf8hpk2kjhhrsdtrf033vn6tc9ebcg"
            }
        }
    }
}
*/
func (tx Tx) DposVotes() (votes []DposVote) {
	switch tx.Typ {
	case TxTypeCertification, TxTypeGenesis, TxTypeWork: //这几类不会影响投票
		return
	default:
	}

	appendVote := func(delegate, voter string, amount decimal.Decimal) {
		votes = append(votes, DposVote{
			BlockHeight: tx.BlockHeight,
			Txid:        tx.Txid,
			Delegate:    delegate,
			Voter:       voter,
			Amount:      amount,
		})
	}

	amount, negAmount := tx.Amount, tx.Amount.Add(tx.Txfee).Neg()
	from, to := tx.SendFrom, tx.SendTo
	fromPrefix, toPrefix := tx.SendFrom[:4], tx.SendTo[:4]

	switch fromPrefix {
	case delegateAddrPrefix_20m0:
		appendVote(from, from, negAmount)
		switch toPrefix {
		case delegateAddrPrefix_20m0:
			appendVote(to, to, amount)
		case voteAddrPrefix_20w0:
			d, v := getVote(tx.Sig[:132])
			appendVote(d, v, amount)
		default:
		}
	case voteAddrPrefix_20w0:
		var fromAddrTpl string
		switch toPrefix {
		case delegateAddrPrefix_20m0:
			fromAddrTpl = tx.Sig[:132]
			appendVote(to, to, amount)
		case voteAddrPrefix_20w0:
			d, v := getVote(tx.Sig[:132])
			appendVote(d, v, amount)
			fromAddrTpl = tx.Sig[132:264]
		default:
			fromAddrTpl = tx.Sig[:132]
		}
		d, v := getVote(fromAddrTpl)
		appendVote(d, v, negAmount)
	default:
		switch toPrefix {
		case delegateAddrPrefix_20m0:
			appendVote(to, to, amount)
		case voteAddrPrefix_20w0:
			d, v := getVote(tx.Sig[132:])
			appendVote(d, v, amount)
		default:
		}
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
