# TinyRedis

## 功能

1. TCP服务端
    - 基于 go netpoller 实现 IO 多路复用
    - 支持 RESP 协议
2. 数据类型与操作指令
    - string——get/mget/set/mset
    - list——lpush/lpop/rpush/rpop/lrange
    - set——sadd/sismember/srem
    - hashmap——hset/hget/hdel
    - sortedset——zadd/zrem/zrangebyscore
3. 数据持久化
    - appendonlyfile落盘与重写

##  Benchmark

```
SET: 137362.64 requests per second, p50=0.207 msec                    
GET: 239234.44 requests per second, p50=0.111 msec                    
LPUSH: 129870.13 requests per second, p50=0.207 msec                    
RPUSH: 125944.58 requests per second, p50=0.215 msec                    
LPOP: 119617.22 requests per second, p50=0.223 msec                    
RPOP: 121654.50 requests per second, p50=0.223 msec                    
SADD: 122699.39 requests per second, p50=0.215 msec                    
HSET: 123915.74 requests per second, p50=0.223 msec                    
ZADD: 122399.02 requests per second, p50=0.223 msec                    
LPUSH (needed to benchmark LRANGE): 119904.08 requests per second, p50=0.223 msec                    
LRANGE_100 (first 100 elements): 95238.10 requests per second, p50=0.327 msec                    
LRANGE_300 (first 300 elements): 54734.54 requests per second, p50=0.535 msec                   
LRANGE_500 (first 500 elements): 32499.19 requests per second, p50=0.695 msec                   
LRANGE_600 (first 600 elements): 27270.25 requests per second, p50=0.791 msec                   
MSET (10 keys): 95147.48 requests per second, p50=0.303 msec
```

### 单类命令（50并发，3B）

| 命令 | 吞吐 | 平均延迟 | P50 | P95 | P99 | 最大 |
|---|---:|---:|---:|---:|---:|---:|
| SET | 91,743 | 0.363 | 0.271 | 0.751 | 2.431 | 4.351 |
| GET | 188,679 | 0.168 | 0.119 | 0.239 | 0.663 | 12.991 |
| MSET(10 keys) | 100,000 | 0.437 | 0.447 | 0.583 | 0.607 | 0.631 |
| LPUSH | 114,943 | 0.282 | 0.223 | 0.383 | 1.791 | 7.935 |
| RPUSH | 119,048 | 0.284 | 0.231 | 0.583 | 1.415 | 1.847 |
| LPOP | 117,647 | 0.275 | 0.223 | 0.559 | 1.367 | 2.015 |
| RPOP | 121,951 | 0.298 | 0.191 | 0.487 | 3.855 | 7.967 |
| SADD | 102,041 | 0.333 | 0.239 | 0.551 | 3.295 | 4.543 |
| HSET | 114,943 | 0.250 | 0.231 | 0.311 | 0.607 | 3.487 |
| ZADD | 108,696 | 0.356 | 0.231 | 0.543 | 2.431 | 15.423 |


### 混合操作（100并发，3B）

| 命令 | 吞吐 | 平均延迟 | P50 | P95 | P99 | 最大 |
|---|---:|---:|---:|---:|---:|---:|
| SET | 147,059 | 0.467 | 0.455 | 0.663 | 0.711 | 0.831 |
| GET | 172,414 | 0.297 | 0.271 | 0.439 | 0.487 | 0.703 |
| LPUSH | 94,340 | 0.628 | 0.623 | 1.167 | 1.367 | 1.495 |
| SADD | 178,571 | 0.349 | 0.335 | 0.591 | 0.671 | 0.815 |
| HSET | 104,167 | 0.569 | 0.527 | 1.415 | 1.743 | 1.879 |
| ZADD | 72,464 | 0.924 | 0.959 | 1.431 | 1.719 | 1.999 |


### 不同并发（SET/GET，3B）

| 并发 | SET 吞吐 | SET 平均延迟 | GET 吞吐 | GET 平均延迟 |
|---:|---:|---:|---:|---:|
| 10 | 142,857 | 0.047 | 104,167 | 0.074 |
| 200 | 86,207 | 1.444 | 172,414 | 0.591 |


### 不同负载大小（SET/GET，50并发）

| 负载 | SET 吞吐 | SET 平均延迟 | GET 吞吐 | GET 平均延迟 |
|---:|---:|---:|---:|---:|
| 64B | 98,039 | 0.279 | 156,250 | 0.170 |
| 1KB | 142,857 | 0.203 | 200,000 | 0.134 |





