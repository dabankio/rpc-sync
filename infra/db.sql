-- 脚本用于单元测试，生产手工建表;
-- create user bbcrpc_sync_usr with password 'pwd';
-- drop database bbcrpc_sync;
-- create database bbcrpc_sync with owner bbcrpc_sync_usr;

create table if not exists blocks (
    height integer primary key,
    hash text not null,
    prev_hash text,
    version smallint,
    typ text not null,
    time timestamp with time zone not null,
    fork text not null,
    coinbase numeric(12,6) not null,
    miner text not null,
    tx_count smallint not null
) without oids;

create index if not exists idx_block_time on blocks using btree (time);

--txs用 block_height 做表分区以避免单表过度膨胀，txs访问必须指定block_height (可以结合blocks表使用),以50w高度为一个分区， txs_1 存储1~50w高度的数据, txs_2 50w~100w;

create table txs (
    block_height integer REFERENCES blocks(height),
    txid text not null,
    version integer,
    typ text not null,
    time timestamp with time zone not null,
    lockuntil integer,
    anchor text,
    block_hash text,
    send_from text,
    send_to text,
    amount numeric check(amount >= 0),
    txfee numeric check(txfee >= 0),
    data text,
    sig text,
    fork text,
    vin jsonb,
    CONSTRAINT unq_height_txid UNIQUE (block_height, txid)
) PARTITION BY RANGE (block_height) without oids;

create index on txs (block_height);

CREATE TABLE txs_1 PARTITION OF txs FOR VALUES FROM (0) TO (500000);
CREATE TABLE txs_2 PARTITION OF txs FOR VALUES FROM (500000) TO (1000000);
CREATE TABLE txs_3 PARTITION OF txs FOR VALUES FROM (1000000) TO (1500000);
CREATE TABLE txs_4 PARTITION OF txs FOR VALUES FROM (1500000) TO (2000000);
CREATE TABLE txs_5 PARTITION OF txs FOR VALUES FROM (2000000) TO (2500000);
CREATE TABLE txs_6 PARTITION OF txs FOR VALUES FROM (2500000) TO (3000000);
CREATE TABLE txs_7 PARTITION OF txs FOR VALUES FROM (3000000) TO (3500000);
CREATE TABLE txs_8 PARTITION OF txs FOR VALUES FROM (3500000) TO (4000000);
CREATE TABLE txs_9 PARTITION OF txs FOR VALUES FROM (4000000) TO (4500000);

create table dpos_vote (
    block_height integer REFERENCES blocks(height),
    txid text not null,
    delegate text not null,
    voter text not null,
    amount numeric not null
) without oids;

create table vote_sum ( --中间表，数据可按高度(block_height)删除
    block_height integer REFERENCES blocks(height),
    last_height integer not null,
    delegate text not null,
    voter text not null,
    amount numeric check(amount >= 0),
    CONSTRAINT unq_at_height unique(block_height, delegate, voter),
    CONSTRAINT last_height_let_block_height check(last_height <= block_height)
) without oids;

create table day_reward (
    day date not null,
    delegate text not null,
    voter text not null,
    amount numeric,
    CONSTRAINT unq_day_delegate_voter unique(day, delegate, voter)
) without oids;

create table unblocked_block (
    addr_from text not null,
    addr_to text not null,
    balance numeric,
    time_span integer,
    day date,
    height integer,
    CONSTRAINT unq_from_to_day unique(addr_from, addr_to, day)
)without oids;

--alter table txs owner to bbcrpc_sync_usr;
--alter table dpos_vote owner to bbcrpc_sync_usr;
--alter table vote_sum owner to bbcrpc_sync_usr;
--alter table day_reward owner to bbcrpc_sync_usr;
--alter table unblocked_block owner to bbcrpc_sync_usr;
--alter table blocks owner to bbcrpc_sync_usr;
--alter table txs_1 owner to bbcrpc_sync_usr;
--alter table txs_2 owner to bbcrpc_sync_usr;
--alter table txs_3 owner to bbcrpc_sync_usr;
--alter table txs_4 owner to bbcrpc_sync_usr;
--alter table txs_5 owner to bbcrpc_sync_usr;
--alter table txs_6 owner to bbcrpc_sync_usr;
--alter table txs_7 owner to bbcrpc_sync_usr;
--alter table txs_8 owner to bbcrpc_sync_usr;
--alter table txs_9 owner to bbcrpc_sync_usr;