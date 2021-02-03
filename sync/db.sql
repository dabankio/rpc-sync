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
    tx_count smallint not null,
) without oids;

create table txs (
    block_height integer REFERENCES blocks(height),
    txid text not null,
    version smallint,
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