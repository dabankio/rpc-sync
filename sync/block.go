package sync

import (
	"time"

	"github.com/dabankio/bbrpc"
)

type ConsensusType string

const (
	ConsensusTypePow     ConsensusType = "primary-pow"
	ConsensusTypeDpos    ConsensusType = "primary-dpos"
	ConsensusTypeGenesis ConsensusType = "genesis"
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

type HeightBlockMap map[uint64]Block

func NewHeightBlockMap(blocks []Block) HeightBlockMap {
	m := make(HeightBlockMap, len(blocks))
	for i := 0; i < len(blocks); i++ {
		m[blocks[i].Height] = blocks[i]
	}
	return m
}
