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

**获取leader的过程**

假设集群刚被初始化，当前没有任何节点处于leader节点状态。

我们将列举节点1获得leader状态的过程。