#!/usr/bin/python3
# -*- coding: UTF-8 -*-

import time
import uuid
import requests
import json
import sys
from binascii import hexlify, unhexlify
import struct
import hashlib

url = 'http://127.0.0.1:9902'

str_ = '表决模板功能Beta测试%表决模板功能正在进行Beta测试%A、测试A选项%B、测试B选项%本期表决将于高度599720停止采集投票选项%表决权重快照以于高度592443完成'.encode('utf-8')

s = hashlib.sha256()
s.update(str_)
b = s.hexdigest()

data_json = {
    "id":1,
    "method":"getpubkeyaddress",
    "jsonrpc":"2.0",
    "params": {
        "pubkey":s.hexdigest()
    }
}

response = requests.post(url, json=data_json)
res = json.loads(response.text)

data_json = {
    "id":1,
    "method":"addnewtemplate",
    "jsonrpc":"2.0",
    "params":{
        "type":"dexbbcmap",
        "dexbbcmap":{
            "owner":res["result"]
            }
        }
    }
response = requests.post(url, json=data_json)
res = json.loads(response.text)
print("投票地址:",res["result"])

str_16 = hexlify(str_)
str_time = hexlify(struct.pack("<I", int(time.time())))
data = uuid.uuid1().hex.encode() + str_time + b"00" + str_16

data_json = {
    "id":1,
    "method":"sendfrom",
    "jsonrpc":"2.0",
    "params":{
        "from":"1j6e2a67amk97x5wjb203vxv4hhwqeh47k5zqg4vkhxm01045v5wv805q",
        "to":res["result"],
        "amount":1.00000000,
        "data":data.decode('utf-8')
        }
    }
response = requests.post(url, json=data_json)
res = json.loads(response.text)
data_json = {
    "id":1,
    "method":"gettransaction",
    "jsonrpc":"2.0",
    "params":{
        "txid":res["result"]
        }
    }
response = requests.post(url, json=data_json)
res = json.loads(response.text)
data_str = res["result"]["transaction"]["data"]

print("uuid:", data_str[0:8] + "-" + data_str[8:12] + "-" + data_str[12:16] + "-" + data_str[16:20] + "-" + data_str[20:32])
v = struct.unpack("<I", unhexlify(data_str[32:40]))
timeArray = time.localtime(v[0])
otherStyleTime = time.strftime("%Y-%m-%d %H:%M:%S", timeArray)
print("时间:", otherStyleTime)
print("立项内容:", unhexlify(data_str[42:]).decode('utf-8'))

'''
投票地址 21c0czt5zxc2rgsvfssqa3f5768np6c9rfycqqwzs51jbyhms0b51ek0g
uuid 2339d764-5093-11eb-a9ed-0492265b8709
时间: 2021-01-07 10:50:44
立项内容: 表决模板功能Beta测试%表决模板功能正在进行Beta测试%A、测试A选项%B、测试B选项%本期表决将于高度599720停止采集投票选项%表决权重快照以于高度592443完成

txid: 5ff67704aa68d1872f0d2559b7ec57150ad851ee85bc4c4f41d26cbf8639bb94
'''
