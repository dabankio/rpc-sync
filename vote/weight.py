#!/usr/bin/python
# -*- coding: UTF-8 -*-

import requests
import json
import pymysql
import time, datetime
from decimal import Decimal
import sys
import os
sys.path.append("..")
import config

connection = pymysql.connect(host=config.host, port=config.port, user=config.user, password=config.password, db=config.db)

def Delete():
    with connection.cursor() as cursor:
        cursor.execute("delete from Vote;")
        connection.commit()
    print("Delete OK")

def Insert(max_id):
    with connection.cursor() as cursor:
        sql = "INSERT Vote(addr,a1) SELECT `to` as addr,SUM(amount) as a1 from Tx where left(`to`,4) != '20w0' and id <= %d GROUP BY `to`" % max_id
        cursor.execute(sql)
        connection.commit()
    print("Insert OK")

def Update(max_id):
    with connection.cursor() as cursor:
        sql = "SELECT form,SUM(amount+free) as a2 from Tx where left(form,4) != '20w0' and form != '000000000000000000000000000000000000000000000000000000000' and id <= %d GROUP BY form" % max_id
        cursor.execute(sql)
        rows = cursor.fetchall()
        A2 = {}
        for r in rows:
            A2[r[0]] = r[1]
        print("out amout...")
    
        sql = "SELECT client_in,SUM(amount) as v1 from Tx where client_in is not null and id <= %d GROUP BY client_in" % max_id
        cursor.execute(sql)
        rows = cursor.fetchall()
        V1 = {}
        for r in rows:
            V1[r[0]] = r[1]
        print("in vote...")
    
        sql = "SELECT client_out,SUM(amount+free) as v2 from Tx where client_out is not null and id <= %d GROUP BY client_out" % max_id
        cursor.execute(sql)
        rows = cursor.fetchall()
        V2 = {}
        for r in rows:
            V2[r[0]] = r[1]
        print("out vote...")
    
        cursor.execute("select * from Vote")
        rows = cursor.fetchall()
        for r in rows:
            a2 = 0
            if r[1] in A2:
                a2 = A2[r[1]]

            v1 = 0
            if r[1] in V1:
                v1 = V1[r[1]]

            v2 = 0
            if r[1] in V2:
                v2 = V2[r[1]]
            sql = "update Vote set a2 = %s,v1 = %s,v2 = %s where id = %s"
            cursor.execute(sql,[a2,v1,v2,r[0]])
        connection.commit()
        print("update OK")

def Export(height):
    data = []
    f = open(('./%d.json' % height), 'w')    
    with connection.cursor() as cursor:
        cursor.execute("select addr,(a1 - a2) as a,(v1 - v2) as v from Vote")
        rows = cursor.fetchall()
        for r in rows:
            if r[1] > 0 or r[2] > 0:
                data.append({
                    "address": r[0],
                    "balance": float(r[1]),
                    "vote": float(r[2])})
            if r[1] < 0 or r[2] < 0:
                print("err:",r)
                sys.exit()
    f.write(json.dumps(data))
    f.close()
    print("Export OK")

def GetID(height):
    with connection.cursor() as cursor:
        sql = "select hash from Block where height = %d and is_useful = 1" % height
        cursor.execute(sql)
        sql = "select max(id) as id from Tx where block_hash = '%s'" % cursor.fetchone()[0]
        cursor.execute(sql)
        return cursor.fetchone()[0]
        

if __name__ == '__main__':
    height = 592443
    max_id = GetID(height)
    Delete()
    Insert(max_id)
    Update(max_id)
    Export(height)