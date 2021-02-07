# 奖励计算

## dpos 奖励计算算法

计算时间 T1 到 T2, 对应高度 H1 H2 之间的奖励

H1 之前的一个块为H0

取得截止H0的投票数据汇总（各地址投给各节点的票数总计

取 H1 ~ H2 间的所有块
取 H1 ~ H2 间的所有投票记录

逐个块的计算奖励

在高度H, 矿工M dpos出块，假定高度 H 前一个块的高度为 H_1 `用户在高度 H 的块的奖励 = (截止H_1投给M的票数 / 截止 H_1 M获得的总票数) * 出块奖励 `

再处理下小数位数 （单个块奖励截断至小数位后9位， 单日汇总截断至小数位后6位）


## 新旧数据校验

```sql
--curl https://edposapi.bbcpool.io/api/Reward?date=2021-02-04

create table api_rwd (day date, delegate text,voter text, amount numeric) without oids;

select dr.*, api.amount api_amount, dr.amount - api.amount offset_amount, (dr.amount - api.amount)/(api.amount+0.0001) percent
from day_reward dr
left join api_rwd api on api.delegate = dr.delegate and api.voter = dr.voter and api.day=dr.day
order by dr.amount desc;

with x as (
    select dr.*, api.amount api_amount, dr.amount - api.amount offset_amount
    from day_reward dr
    left join api_rwd api on api.delegate = dr.delegate and api.voter = dr.voter and api.day=dr.day
)
select (x.delegate=x.voter)::bool is_self, x.* from x where x.offset_amount > 0
order by is_self;

insert into api_rwd (day, delegate, voter, amount) values
```