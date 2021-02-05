# 奖励计算

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