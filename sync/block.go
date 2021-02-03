package sync

import (
	"time"

	"github.com/dabankio/bbrpc"
)

type ConsensusType string

const (
	ConsensusTypePow     ConsensusType = "pow"     //primary-pow
	ConsensusTypeDpos    ConsensusType = "dpos"    //primary-dpos
	ConsensusTypeGenesis ConsensusType = "genesis" //genesis
)

type Block struct {
	Height   uint64
	Hash     string
	PrevHash string
	Version  int
	Typ      ConsensusType
	Time     time.Time
	Fork     string
	Coinbase float64
	Miner    string

	TxCount int
}

func NewBlock(blc *bbrpc.BlockDetail) Block {
	return Block{
		Height:   uint64(blc.Height),
		Hash:     blc.Hash,
		PrevHash: blc.Prev,
		Version:  int(blc.Version),
		Typ:      ConsensusType(blc.Type),
		Time:     time.Unix(int64(blc.Time), 0),
		Fork:     blc.Fork,
		TxCount:  len(blc.Tx),
		Coinbase: blc.Txmint.Amount,
		Miner:    blc.Txmint.Sendto,
	}
}
