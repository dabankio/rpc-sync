#!/usr/bin/env python
# -*- coding: UTF-8 -*-

from TokenDistribution import TokenDistribution

td = TokenDistribution()

def getm(v2):
    if v2 > 129600:
        m = 1153 * 43200 + 1043 * 43200  + 933 * 43200 + 823 * (v2 - 129600)
    elif v2 > 86400:
        m = 1153 * 43200 + 1043 * 43200  + 933 * (v2 - 86400)
    elif v2 > 43200:
        m = 1153 * 43200 + 1043 * (v2 - 43200)
    else:
        m = 1153 * v2
    return m

err = 0
for i in range(129800):
    v1 = getm(i)
    v2 = td.GetTotal(i)
    if v1 != v2:
        err = err + 1
        print "err"
print "err=%d" % err