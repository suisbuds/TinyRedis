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
SET: 292397.66 requests per second, p50=0.087 msec                    
GET: 301204.84 requests per second, p50=0.087 msec                    
LPUSH: 299401.22 requests per second, p50=0.087 msec                    
RPUSH: 297619.06 requests per second, p50=0.087 msec                    
LPOP: 302114.81 requests per second, p50=0.087 msec                    
RPOP: 298507.47 requests per second, p50=0.087 msec                    
SADD: 294117.66 requests per second, p50=0.087 msec                    
HSET: 294985.25 requests per second, p50=0.087 msec                    
ZADD: 295858.00 requests per second, p50=0.087 msec                    
LPUSH (needed to benchmark LRANGE): 298507.47 requests per second, p50=0.087 msec                    
LRANGE_100 (first 100 elements): 169204.73 requests per second, p50=0.159 msec                    
LRANGE_300 (first 300 elements): 72992.70 requests per second, p50=0.343 msec                   
LRANGE_500 (first 500 elements): 46620.05 requests per second, p50=0.535 msec                   
LRANGE_600 (first 600 elements): 40436.71 requests per second, p50=0.615 msec                   
MSET (10 keys): 262467.19 requests per second, p50=0.159 msec   
```

### 单类命令（50并发，3B）

| 命令 | 吞吐 | 平均延迟 | P50 | P95 | P99 | 最大 |
|---|---:|---:|---:|---:|---:|---:|
| SET | 96,154 | 0.342 | 0.247 | 1.135 | 1.247 | 1.455 |
| GET | 263,158 | 0.118 | 0.095 | 0.127 | 0.543 | 10.687 |
| MSET(10 keys) | 111,111 | 0.415 | 0.407 | 0.879 | 1.095 | 1.183 |
| LPUSH | 161,290 | 0.216 | 0.175 | 0.423 | 0.871 | 1.327 |
| RPUSH | 120,482 | 0.308 | 0.183 | 0.647 | 1.751 | 15.487 |
| LPOP | 163,934 | 0.193 | 0.175 | 0.327 | 0.583 | 1.159 |
| RPOP | 181,818 | 0.202 | 0.191 | 0.343 | 0.447 | 1.191 |
| SADD | 158,730 | 0.206 | 0.191 | 0.351 | 0.471 | 1.079 |
| HSET | 166,667 | 0.207 | 0.191 | 0.367 | 0.423 | 0.791 |
| ZADD | 161,290 | 0.224 | 0.167 | 0.351 | 0.839 | 10.879 |


### 混合操作（100并发，3B）

| 命令 | 吞吐 | 平均延迟 | P50 | P95 | P99 | 最大 |
|---|---:|---:|---:|---:|---:|---:|
| SET | 208,333 | 0.347 | 0.351 | 0.535 | 0.639 | 0.711 |
| GET | 185,185 | 0.270 | 0.207 | 0.615 | 0.847 | 1.119 |
| LPUSH | 119,048 | 0.667 | 0.447 | 1.703 | 7.775 | 7.903 |
| SADD | 121,951 | 0.567 | 0.543 | 1.023 | 1.183 | 1.239 |
| HSET | 142,857 | 0.510 | 0.359 | 1.191 | 1.879 | 9.535 |
| ZADD | 131,579 | 0.437 | 0.367 | 0.839 | 1.063 | 1.271 |


### 不同并发（SET/GET，3B）

| 并发 | SET 吞吐 | SET 平均延迟 | GET 吞吐 | GET 平均延迟 |
|---:|---:|---:|---:|---:|
| 10 | 156,250 | 0.046 | 238,095 | 0.029 |
| 200 | 94,340 | 1.720 | 208,333 | 0.501 |


### 不同负载大小（SET/GET，50并发）

| 负载 | SET 吞吐 | SET 平均延迟 | GET 吞吐 | GET 平均延迟 |
|---:|---:|---:|---:|---:|
| 64B | 238,095 | 0.141 | 227,273 | 0.112 |
| 1KB | 125,000 | 0.281 | 250,000 | 0.112 |





