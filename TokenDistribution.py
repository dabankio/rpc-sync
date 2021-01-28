#!/usr/bin/python
# -*- coding: UTF-8 -*-
 
class TokenDistribution:
    def __init__(self):
        self.m_ltd =  []
        total = 0
        height = 0
        for i in range(5):
            reward = 1153 - 110 * i
            total += 43200 * reward
            height += 43200
            self.m_ltd.append({"reward":reward,"total":total,"height":height})
    
        for i in range(11):
            reward = 603 - 53 * i

            total += (43200 * 5) * reward
            height += (43200 * 5)
            self.m_ltd.append({"reward":reward,"total":total,"height":height})

    def GetTotal(self,height):
        max_height = 0
        max_money = 0
        for obj in self.m_ltd:
            if height < obj["height"]:
                return obj["total"] - (obj["height"] - height) * obj["reward"]
            max_height = obj["height"]
            max_money = obj["total"]
        return max_money + (height - max_height) * 20

    def PrintInfo(self):
        for obj in self.m_ltd:
            print(obj)