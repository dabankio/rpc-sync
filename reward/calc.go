package reward

import (
	"bbcsyncer/sync"
	"log"
	"time"

	"github.com/dabankio/civil"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const minVoteHeight = 243802 //首次dpos投票的高度
const minDposHeight = 243973 //首次dpos出块的高度，低于这个高度的奖励数据不统计
const bbcDecimals__ = 6

var ZoneBeijingTime = time.FixedZone("Beijing", 8*int(time.Hour.Seconds()))

func NewCalc(repo *Repo, syncRepo *sync.Repo) *Calc {
	return &Calc{repo: repo, syncRepo: syncRepo}
}

type Calc struct {
	repo     *Repo
	syncRepo *sync.Repo
}

func (c *Calc) CalcAtDayEast8zoneAndSave(day civil.Date) ([]DayReward, error) {
	rwds, err := c.CalcAtDayEast8zone(day)
	if err != nil {
		return nil, err
	}
	err = c.repo.InsertDayReward(rwds)
	return rwds, err
}
func (c *Calc) CalcAtDayEast8zone(day civil.Date) ([]DayReward, error) {
	defer func(startAt time.Time) {
		log.Println("calc_at_day_done, cost: ", time.Now().Sub(startAt))
	}(time.Now())

	fromTime := day.In(ZoneBeijingTime)
	toTime := fromTime.Add(24 * time.Hour).Add(-time.Microsecond)
	rewards, err := c.Calc(fromTime, toTime)
	if err != nil {
		return nil, err
	}
	var items []DayReward
	for _, rwd := range rewards {
		items = append(items, DayReward{
			Day:      day,
			Delegate: rwd.Delegate,
			Voter:    rwd.Voter,
			Amount:   rwd.Amount,
		})
	}
	return items, nil
}

//统计时间范围（含端点）内的dpos奖励
func (c *Calc) Calc(fromTime, toTime time.Time) ([]Reward, error) {
	fromHeight, toHeight, err := c.repo.BlockHeightBetween(fromTime, toTime)
	if err != nil {
		return nil, errors.Wrap(err, "get blocks height between err")
	}
	if fromHeight < minDposHeight {
		return nil, errors.Errorf("高度小于 %d 是不统计dpos奖励", minDposHeight)
	}

	blocks, err := c.syncRepo.BlocksBetweenHeight(fromHeight, toHeight)
	if err != nil {
		return nil, errors.Wrap(err, "get blocks between height err")
	}
	prevFromHeightVoteSums, err := c.voteSumAtHeight(fromHeight - 1) //截止前一个块时的投票汇总
	if err != nil {
		return nil, errors.Wrap(err, "get vote sum err")
	}
	votes, err := c.syncRepo.DposVotesBetweenHeight(fromHeight, toHeight)
	if err != nil {
		return nil, errors.Wrap(err, "get votes err")
	}

	return calcRewards(blocks, prevFromHeightVoteSums, votes, fromHeight, toHeight)
}

type MinerAddress string

func calcRewards(blocks []sync.Block, sums []VoteSum, votes []sync.DposVote, fromHeight, toHeight uint64) ([]Reward, error) {
	if len(blocks) == 0 {
		return nil, errors.New("no blocks")
	}
	if blocks[0].Height != fromHeight || blocks[len(blocks)-1].Height != toHeight {
		return nil, errors.New("blocks 似乎和需要统计的高度范围不一致")
	}
	if len(blocks) != int(toHeight-fromHeight)+1 {
		return nil, errors.Errorf("blocks len (%d) not valid, want: %d", len(blocks), toHeight-fromHeight+1)
	}

	voteMap := make(map[MinerAddress]map[string]decimal.Decimal) //key: delegate, value: map[voter]amount
	for _, sum := range sums {
		del := MinerAddress(sum.Delegate)
		if _, ok := voteMap[del]; !ok {
			voteMap[del] = make(map[string]decimal.Decimal)
		}
		voteMap[del][sum.Voter] = sum.Amount
	}

	votesAtHeightMap := make(map[uint64][]sync.DposVote) //key: blockHeight, value: votes
	for _, vote := range votes {
		votesAtHeightMap[vote.BlockHeight] = append(votesAtHeightMap[vote.BlockHeight], vote)
	}

	rewardMap := make(map[MinerAddress]map[string]decimal.Decimal)

	lastHeight := blocks[0].Height - 1
	for _, block := range blocks {
		if block.Height < fromHeight || block.Height > toHeight {
			return nil, errors.Errorf("block height (%d) out of range (%d-%d)", block.Height, fromHeight, toHeight)
		}
		if block.Height != lastHeight+1 {
			return nil, errors.Errorf("区块似乎不连续 last: %d, this: %d", lastHeight, block.Height)
		}

		coinbaseAmount := decimal.NewFromFloat(block.Coinbase)
		miner := MinerAddress(block.Miner)
		lastHeight = block.Height
		if block.Typ == sync.ConsensusTypeDpos {
			totalVoteAmount := decimal.NewFromInt(0)
			for _, voteAmount := range voteMap[miner] {
				totalVoteAmount = totalVoteAmount.Add(voteAmount)
			}
			for voter, voteAmount := range voteMap[miner] {
				rewardAmountAtHeight := voteAmount.
					Div(totalVoteAmount).
					Mul(coinbaseAmount).
					Truncate(bbcDecimals__ + 3)

				if _, ok := rewardMap[miner]; !ok {
					rewardMap[miner] = make(map[string]decimal.Decimal)
				}
				if _, ok := rewardMap[miner][voter]; !ok {
					rewardMap[miner][voter] = decimal.NewFromInt(0)
				}
				rewardMap[miner][voter] = rewardMap[miner][voter].Add(rewardAmountAtHeight)
				log.Printf("[dbg] height: %d, delegate: %s, voter: %s, reward: %s \n", block.Height, miner, voter, rewardAmountAtHeight)
			}
		}
		for _, vote := range votesAtHeightMap[block.Height] {
			del := MinerAddress(vote.Delegate)
			if _, ok := voteMap[del]; !ok {
				voteMap[del] = make(map[string]decimal.Decimal)
			}
			if _, ok := voteMap[del][vote.Voter]; !ok {
				voteMap[del][vote.Voter] = decimal.NewFromInt(0)
			}
			voteMap[del][vote.Voter] = voteMap[del][vote.Voter].Add(vote.Amount)
		}
	}

	var rewards []Reward
	zero := decimal.NewFromInt(0)
	for miner, mrm := range rewardMap {
		for voter, rewardAmount := range mrm {
			if rewardAmount.Equal(zero) {
				continue
			}
			rewards = append(rewards, Reward{
				FromHeight: fromHeight,
				ToHeight:   toHeight,
				Delegate:   string(miner),
				Voter:      voter,
				Amount:     rewardAmount.Truncate(bbcDecimals__),
			})
		}
	}
	return rewards, nil
}

func (c *Calc) voteSumAtHeight(height uint64) ([]VoteSum, error) {
	sums, err := c.repo.SelectVoteSumAtHeight(height)
	if err != nil {
		return nil, err
	} else {
		if len(sums) == 0 { //认为还没统计过
			err = c.repo.CreateVoteSumAtHeight(height)
			if err != nil {
				return nil, errors.Wrap(err, "统计区块高度时的投票汇总时报错")
			}
			return c.repo.SelectVoteSumAtHeight(height)
		} else {
			return sums, nil
		}
	}
}
