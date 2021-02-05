package reward

import "github.com/shopspring/decimal"

// VoteSum 截止某一高度(含该高度)的投票情况
type VoteSum struct {
	BlockHeight uint64
	LastHeight  uint64 //这个高度下的最后一次投票
	Delegate    string
	Voter       string
	Amount      decimal.Decimal
}

/**
select delegate, voter, sum(amount) amount, max(block_height) h from dpos_vote
where delegate = '20m03w2c5xhphzfq7fqzh8qfgpdgsn86dzzdrhxb613ar2frg5y71t2yx'
group by delegate, voter
order by amount desc;


insert into vote_sum (block_height, delegate, voter, last_height, amount)
select 300000 block_height, delegate, voter, max(block_height) last_height, sum(amount) amount
from dpos_vote
where block_height <= 300000
group by delegate, voter

*/
