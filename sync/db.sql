-- create user bbcrpc_sync_usr with password 'pwd';
-- drop database bbcrpc_sync;create database bbcrpc_sync with owner bbcrpc_sync_usr;

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

create table txs (
    block_height integer REFERENCES blocks(height),
    txid text not null,
    version integer,
    typ text not null,
    time timestamp with time zone,
    lockuntil integer,
    anchor text,
    block_hash text,
    send_from text,
    send_to text,
    amount numeric,
    txfee numeric,
    data text,
    sig text,
    fork text,
    vin jsonb
) without oids;

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

alter table blocks owner to bbcrpc_sync_usr;
alter table txs owner to bbcrpc_sync_usr;
alter table dpos_vote owner to bbcrpc_sync_usr;
alter table vote_sum owner to bbcrpc_sync_usr;
alter table day_reward owner to bbcrpc_sync_usr;