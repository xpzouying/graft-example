# graft-example

通过示例分解graft底层代码。

## 运行环境

graft运行环境如下，

- nats server：单点。运行在localhost:4222。其中topic有3个，分别是：
  - vote_request：发起投票
  - vote_response：响应投票
  - heartbeat：发送心跳

- graft-example：集群数量为3，RPC通过nats server相互通信，通信协议为protobuf。

**架构**

![init](./0_init.jpg)


## 流程分析

### 初始化

该节主要分析集群初始化后，各类型节点在启动后的一系列操作。

节点启动后，会对各自做初始化操作，**其中最重要的是订阅nats中的2个topic，建立集群节点之间的通信。**

1. **heartbeat：**
   - 接收集群中其他的节点的心跳。
   - 集群所有节点都订阅相同的topic，例如：graft.cluster.heartbeat。

2. **vote_request：**
   - 接收其他节点发送的投票。
   - 集群所有节点都订阅相同的topic，例如：graft.cluster.vote_request。

**初始化 - Candidate状态启动**

假设集群刚被初始化，当前没有任何节点处于Leader节点状态，也即没有收到任何节点的心跳信息。

1. 节点启动后，状态为Follower。
2. 由于当前集群没有Leader，导致当前term的选举超时（election timeout），节点转换为Candidate，同时更新term为term+1。各个节点的超时时间是随机超时时间，假设此时Node1首先超时。

我们将列举Node1获得leader状态的过程：

1. 向集群中发起投票。
   - 投票给自己
   - 通过nats server的cluster.vote_request队列中发出投票。

2. 根据投票结果，进行不同的状态转换，
   - 赢得多数投票：将当前节点更新为Leader节点。
   - 未赢得多数投票。下列详细分析：

发起投票后，在等待投票结果的过程中，会出现下列的情况，

a. 当前选举周期超时，即没有节点获取到多数投票。此时重置状态，重新进行投票选举。

b. 从node1.response中获取其他节点的响应：
   1. 判断response是否有效：是否投票给自己；是否是当前的term投票。
   2. 若response有效，则选票计数+1。
   3. 若获得多数选票，则切换为Leader节点。
   4. 否则，继续等待其他节点的投票结果。

c. 从cluster.vote_request中获取其他节点的投票请求：
   1. 如果发起投票的请求term落后于node1的term，则发送拒绝response；
   2. 如果对于当前term，node1已经投票给其他节点，则发送拒绝response；
   3. 如果node1接收到的其他节点的request中的term大于node1的term，则更新term到更新的term，并重置当前term的信息，包括：投票信息和leader信息。
   4. 如果node1已经是leader，则直接返回拒绝票。
   5. 否则，上述都没有触发。则投票给请求者。

d. 从cluster.heartbeat中获取集群中的心跳信息：
   1. 如果当前节点还没有设定Leader，则设置心跳信息中的Leader为当前节点的Leader。
   2. 如果当前节点已经有设定Leader，则需要根据term的比对进行处理：
      - 当前term > 心跳term，则忽略心跳请求。
      - 心跳term > 当前term，则更新term到心跳term，且重置投票状态，并且更新当前节点的Leader为心跳节点的Leader。
      - 心跳term == 当前term，且当前节点为Leader节点，则忽略。
      - 心跳term == 当前term，且当前节点为Candidate状态，则投票给心跳节点的Leader。
