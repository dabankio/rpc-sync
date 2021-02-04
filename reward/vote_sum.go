package reward

import "github.com/shopspring/decimal"

// VoteSum 截止某一高度(含该高度)的投票情况
type VoteSum struct {
	BlockHeight uint64
	Delegate    string
	Voter       string
	Amount      decimal.Decimal
}

/**
select delegate, voter, sum(amount) amount, max(block_height) h from dpos_vote
where delegate = '20m03w2c5xhphzfq7fqzh8qfgpdgsn86dzzdrhxb613ar2frg5y71t2yx'
group by delegate, voter
order by amount desc;

*/
