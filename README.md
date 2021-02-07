####  
```
protoc ./dbp/dbp.proto --python_out=./
protoc ./dbp/lws.proto --python_out=./
protoc ./dbp/sn.proto --python_out=./
```

#### 
``` sh
sudo docker ps -a
sudo docker exec -it 3a7c83e6b635 /bin/bash
pip install pymysql
pip install protobuf
pip install requests
pip install ed25519
pip install paho.mqtt
pip install msgpack
pip install bson
```


## 同步策略/逻辑

根据高度进行同步，
同步开始前首先检查已经同步的区块有没有包含分叉的区块，如果已经同步的数据包含废弃的（分叉了的）则移除分叉区块
同步至最新高度

对于每一个区块，在事务中写入这些数据：
- 区块原始数据
- 区块内的交易数据
- 区块内交易产生的投票数据

## dpos投票奖励的计算

设计目标：
- 尽可能依靠原始数据进行计算（中间数据不要太多，需要清晰的逻辑
- 计算需要满足一定的性能需求（计算1天的奖励控制在5分钟内
- 计算需要尽可能灵活，支持计算任意时间段（不超过24小时）内的奖励数据（包括细节数据）
- 对于性能的需求设计目标在10年内（10年内不优化依然可以满足性能要求，存在简单的优化策略


奖励计算算法：
对于任意的超级节点 delegate，地址A B C给 delegate 投票

在高度h之前A B C的投票额分别为 a b c, delegate获得的总投票额为 `（a + b + c）`

delegate 在高度 h 完成了dpos出块，区块奖励（含交易手续费）coinbase 个BBC

那么，对于这个区块，A B C 应该获得投票奖励，计算公式如下：
- A的奖励： `coinbase*a/(a+b+c)`
- B的奖励： `coinbase*b/(a+b+c)`


h高度之后，delegate的投票分布为： `A: a, B: b, C: c, delegate: coinbase`, 后续delegate出块时,delegate本身也可以获得投票奖励

在h1高度时，delegate再次dpos出块，区块奖励（含交易手续费）coinbase2 个BBC,对于h1高度的部分奖励计算公式：
- A的奖励： `coinbase2*a/(a+b+c+coinbase)`
- delegate的奖励： `coinbase2*coinbase/(a+b+c+coinbase)`

h1高度之后coinbase2计入delegate自身投票

数据与统计，对于计算节点delegate 在 t1~t1+deltaT 时间区间内的奖励：
- t1前的最近一个块高度为 h0
- 统计截止h0的投票情况，假设：`A: a, B: b, C: c, delegate: del`，从这个点开始算起
- 查询t1~t1+deltaT直接的所有投票数据，对于每个区块 h
    - 基于h0计算h0～h的投票情况，根据投票量瓜分出块奖励
- 区间内的奖励加和的到每个地址获得的奖励



## 奖励结算
以东八区时区为准，计算一天内的出块奖励情况


## 其他计算细节

如何在一笔交易中取得投票信息（投票人、投票给谁）
- dpos出块获得的出块奖励计为超级节点自投
- 如果to from地址都以20w0开头，则用户从一个节点投票撤回转投另一个节点,解析`sig[0:132]`获得投票信息，解析`sig[132:264]`获得撤回投票信息
- 如果to地址以20w0开头，则是一笔投票, 解析`sig[:132]`获得投票信息
- 如果from地址以20w0开头，则是一笔投票撤回,解析`sig[:132]`获得投票信息
- sig信息参考模版原始数据(132, 264这个)

投票金额
- 投票金额等于交易金额
- 撤回投票金额等于交易金额+手续费金额

如何判定一个区块是不是因为分叉被废弃了
- 如果这个高度的当前区块hash与已经同步的不一致，则认为已经同步的区块数据因为分叉被遗弃了


## 数据量参考

- 每天的出块总数大概是 1140 个块，因此产生的节点自投差不多也是这个量，1年大概是46w节点自投记录
- dpos上线大概是2020-05月，截止 2021-02-01 ，大概产生了8w笔独立投票