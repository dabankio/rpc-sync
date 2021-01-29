#!/usr/bin/env python3
# -*- coding: UTF-8 -*-
'''
参考地址 https://github.com/FissionAndFusion/FnFnCoreWallet/wiki/Socket%E6%8E%A5%E5%8F%A3%E5%8D%8F%E8%AE%AE
'''

from socket import *
import dbp_pb2
import lws_pb2
import sys
import struct
from binascii import hexlify, unhexlify
import time
import task
import config

def Run():
    ADDR = (config.dbp_ip,config.dbp_port)
    s = socket(AF_INET,SOCK_STREAM)
    s.connect(ADDR)
    conn = dbp_pb2.Connect()
    conn.session = ""
    conn.version = 1
    conn.client = "lws"
    obj = lws_pb2.ForkID()
    obj.ids.append(config.forkid)
    conn.udata["forkid"].Pack(obj)
    b = dbp_pb2.Base()
    b.msg = dbp_pb2.Msg.CONNECT
    b.object.Pack(conn)
    msg = b.SerializeToString()
    l = struct.pack(">I", len(msg))
    s.send(l + msg)
    ret = s.recv(1024)
    base = dbp_pb2.Base()
    base.ParseFromString(ret[4:])
    if base.msg == dbp_pb2.Msg.FAILED:
        failed = dbp_pb2.Failed()
        base.object.Unpack(failed)
        print("failed:",failed)
        s.close()
        return
    if base.msg == dbp_pb2.Msg.CONNECTED:
        Connected = dbp_pb2.Connected()
        base.object.Unpack(Connected)
        print("Connected:",Connected)

    b.msg = dbp_pb2.Msg.SUB
    sub = dbp_pb2.Sub()
    sub.id = "tx"
    sub.name = "all-tx"
    b.object.Pack(sub)
    msg = b.SerializeToString()
    l = struct.pack(">I", len(msg))
    s.send(l + msg)

    ret = s.recv(1024)
    base = dbp_pb2.Base()
    base.ParseFromString(ret[4:])
    if base.msg == dbp_pb2.Msg.READY:
        ready = dbp_pb2.Ready()
        base.object.Unpack(ready)
        print("ready:",ready)
    if base.msg == dbp_pb2.Msg.NOSUB:
        nosub = dbp_pb2.Nosub()
        base.object.Unpack(nosub)
        print("nosub:",nosub)

    b.msg = dbp_pb2.Msg.SUB
    sub = dbp_pb2.Sub()
    sub.id = "block"
    sub.name = "all-block"
    b.object.Pack(sub)
    msg = b.SerializeToString()
    l = struct.pack(">I", len(msg))
    s.send(l + msg)

    ret_ = s.recv(4)
    l = struct.unpack(">I",ret_)[0]
    ret = s.recv(l)
    base = dbp_pb2.Base()
    base.ParseFromString(ret[4:])
    if base.msg == dbp_pb2.Msg.READY:
        ready = dbp_pb2.Ready()
        base.object.Unpack(ready)
        print("ready:",ready)
    if base.msg == dbp_pb2.Msg.NOSUB:
        nosub = dbp_pb2.Nosub()
        base.object.Unpack(nosub)
        print("nosub:",nosub)

    while True:
        ret = s.recv(1024*1024)
        base = dbp_pb2.Base()
        base.ParseFromString(ret[4:])
        if base.msg == dbp_pb2.Msg.ADDED:
            add = dbp_pb2.Added()
            base.object.Unpack(add)
            if add.id == "tx":
                tx = lws_pb2.Transaction()
                add.object.Unpack(tx)
                txid = hexlify(tx.hash[::-1]).decode()
                task.InsertTxPool(txid)
            if add.id == "block":
                block = lws_pb2.Block()
                add.object.Unpack(block)
                blid = hexlify(block.hash[::-1]).decode()
                #print(blid)
                task.ExecTask(blid)
        
        if base.msg == dbp_pb2.Msg.PING:
            p = dbp_pb2.Pong()
            p.id = "1"
            b.object.Pack(p)
            msg = b.SerializeToString()
            l = struct.pack(">I", len(msg))
            s.send(l + msg)
            print(time.strftime("%H:%M:%S", time.localtime()),"OK")
    s.close()

if __name__ == '__main__':
    Run()
