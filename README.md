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

### 单类命令（50并发，3B）

| 命令 | 吞吐 | 平均延迟 | P50 | P95 | P99 | 最大 |
|---|---:|---:|---:|---:|---:|---:|
| SET | 96,154 | 0.342 | 0.247 | 1.135 | 1.247 | 1.455 |
| GET | 120,482 | 0.222 | 0.207 | 0.311 | 0.399 | 0.887 |
| MSET(10 keys) | 100,000 | 0.366 | 0.303 | 0.615 | 1.143 | 1.271 |
| LPUSH | 38,760 | 1.257 | 1.399 | 2.191 | 2.431 | 2.591 |
| RPUSH | 89,286 | 0.300 | 0.303 | 0.399 | 0.495 | 0.919 |
| LPOP | 129,870 | 0.213 | 0.191 | 0.327 | 0.735 | 1.287 |
| RPOP | 131,579 | 0.209 | 0.191 | 0.311 | 0.527 | 1.039 |
| SADD | 135,135 | 0.217 | 0.183 | 0.559 | 0.647 | 1.039 |
| HSET | 131,579 | 0.212 | 0.191 | 0.335 | 0.615 | 0.983 |
| ZADD | 140,845 | 0.198 | 0.183 | 0.359 | 0.583 | 1.671 |


### 混合操作（100并发，3B）

| 命令 | 吞吐 | 平均延迟 | P50 | P95 | P99 | 最大 |
|---|---:|---:|---:|---:|---:|---:|
| SET | 90,909 | 0.588 | 0.567 | 0.775 | 1.015 | 1.183 |
| GET | 131,579 | 0.413 | 0.367 | 0.591 | 1.391 | 2.175 |
| LPUSH | 34,722 | 2.804 | 2.943 | 4.575 | 4.735 | 5.215 |
| SADD | 87,719 | 0.627 | 0.607 | 0.863 | 0.999 | 1.231 |
| HSET | 121,951 | 0.513 | 0.391 | 1.143 | 1.247 | 1.423 |
| ZADD | 131,579 | 0.437 | 0.367 | 0.839 | 1.063 | 1.271 |


### 不同并发（SET/GET，3B）

| 并发 | SET 吞吐 | SET 平均延迟 | GET 吞吐 | GET 平均延迟 |
|---:|---:|---:|---:|---:|
| 10 | 135,135 | 0.047 | 138,889 | 0.045 |
| 200 | 94,340 | 1.720 | 102,041 | 1.506 |


### 不同负载大小（SET/GET，50并发）

| 负载 | SET 吞吐 | SET 平均延迟 | GET 吞吐 | GET 平均延迟 |
|---:|---:|---:|---:|---:|
| 64B | 131,579 | 0.228 | 131,579 | 0.223 |
| 1KB | 111,111 | 0.317 | 166,667 | 0.195 |


```
SET: 122399.02 requests per second, p50=0.191 msec
GET: 110132.16 requests per second, p50=0.247 msec
LPUSH: 1391.50 requests per second, p50=35.263 msec
RPUSH: 74962.52 requests per second, p50=0.367 msec
LPOP: 82781.46 requests per second, p50=0.311 msec
RPOP: 95419.85 requests per second, p50=0.287 msec
SADD: 86956.52 requests per second, p50=0.303 msec
HSET: 95238.10 requests per second, p50=0.287 msec
ZADD: 87032.20 requests per second, p50=0.303 msec
LPUSH (needed to benchmark LRANGE): 1362.56 requests per second, p50=36.159 msec
LRANGE_100 (first 100 elements): 45351.48 requests per second, p50=0.567 msec
LRANGE_300 (first 300 elements): 22568.27 requests per second, p50=1.095 msec
LRANGE_500 (first 500 elements): 15532.77 requests per second, p50=1.615 msec
LRANGE_600 (first 600 elements): 13049.72 requests per second, p50=1.895 msec
MSET (10 keys): 64432.99 requests per second, p50=0.743 msec
```



