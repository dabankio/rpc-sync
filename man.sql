--日常查询sql

--统计截止某一区块的投票数据
insert into vote_sum (block_height, delegate, voter, last_height, amount)
select 637246 block_height, delegate, voter, max(block_height) last_height, sum(amount) amount
from dpos_vote
where block_height <= 637246
group by delegate, voter;


--查询截止某一区块的投票数据
select * from vote_sum 
where block_height=637246 
and delegate='20m0emvkr82b7qn8hq2fc8tt2135tph52ghnkqs8ggm06qfn42d6zxheq' 
and amount > 0 
order by last_height desc;

--查询截止某一区块的超级节点获得投票情况
select delegate, sum(amount) amount from vote_sum
where block_height = 637246
group by delegate
order by amount;