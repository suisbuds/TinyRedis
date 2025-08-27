set -e

SERVER_HOST="127.0.0.1"
SERVER_PORT="6379"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    printf "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}\n"
}

error() {
    printf "${RED}[ERROR] $1${NC}\n"
}

info() {
    printf "${BLUE}[INFO] $1${NC}\n"
}

cleanup() {
    info "清理测试环境"
    {
        pkill -f "go run ./main.go" 2>/dev/null || true
        pkill -f "TinyRedis" 2>/dev/null || true
        lsof -ti:6379 2>/dev/null | xargs kill -9 2>/dev/null || true
    } >/dev/null 2>&1
    sleep 1
}

start_server() {
    log "启动 TinyRedis 服务器"
    cleanup
    
    if [ ! -f redis.conf ]; then
        error "redis.conf 配置文件不存在！"
        exit 1
    fi
    nohup go run ./main.go > tests/server.log 2>&1 &
    SERVER_PID=$!
    
    info "等待服务器启动"
    for i in {1..30}; do
        if redis-cli -h $SERVER_HOST -p $SERVER_PORT ping >/dev/null 2>&1; then
            log "服务器启动成功 (PID: $SERVER_PID)"
            return 0
        fi
        sleep 1
    done
    
    error "服务器启动失败！"
    exit 1
}

redis_cmd() {
    redis-cli -h $SERVER_HOST -p $SERVER_PORT "$@"
}

test_strings() {
    log "测试字符串操作"
    
    info "测试 SET/GET"
    redis_cmd SET test_key "test_value" >/dev/null
    result=$(redis_cmd GET test_key)
    [ "$result" = "test_value" ] || { error "SET/GET测试失败: got '$result'"; return 1; }
    
    info "测试 SET NX"
    unique_key="nx_test_$(date +%s%N)"
    result=$(redis_cmd SET "$unique_key" "value1" NX)
    [ "$result" = "1" ] || { error "SET NX首次设置失败: got '$result'"; return 1; }
    
    result=$(redis_cmd SET "$unique_key" "value2" NX)
    [ "$result" = "" ] || { error "SET NX重复设置测试失败: got '$result'"; return 1; }
    
    info "测试 SET EX"
    redis_cmd SET ex_key "expiring_value" EX 2 >/dev/null
    result=$(redis_cmd GET ex_key)
    [ "$result" = "expiring_value" ] || { error "SET EX测试失败: got '$result'"; return 1; }
    
    info "测试 MSET/MGET"
    redis_cmd MSET key1 "value1" key2 "value2" key3 "value3" >/dev/null
    result=$(redis_cmd MGET key1 key2 key3 | tr '\n' ' ')
    [[ "$result" =~ "value1" && "$result" =~ "value2" && "$result" =~ "value3" ]] || { error "MSET/MGET测试失败: got '$result'"; return 1; }
    
    log "字符串操作测试通过"
}

test_lists() {
    log "测试列表操作"
    
    info "测试 LPUSH/LRANGE"
    list_key="list_test_$(date +%s%N)"
    redis_cmd LPUSH "$list_key" "item3" "item2" "item1" >/dev/null
    result=$(redis_cmd LRANGE "$list_key" 0 -1 | tr '\n' ' ')
    [[ "$result" =~ "item1" && "$result" =~ "item2" && "$result" =~ "item3" ]] || { error "LPUSH/LRANGE测试失败: got '$result'"; return 1; }
    
    info "测试 RPUSH"
    redis_cmd RPUSH "$list_key" "item4" >/dev/null
    result=$(redis_cmd LRANGE "$list_key" 0 -1 | tail -1)
    [ "$result" = "item4" ] || { error "RPUSH测试失败: got '$result'"; return 1; }
    
    info "测试 LPOP/RPOP"
    lpop_result=$(redis_cmd LPOP "$list_key")
    rpop_result=$(redis_cmd RPOP "$list_key")
    [ "$lpop_result" = "item1" ] && [ "$rpop_result" = "item4" ] || { error "LPOP/RPOP测试失败: lpop='$lpop_result', rpop='$rpop_result'"; return 1; }
    
    log "列表操作测试通过"
}

test_sets() {
    log "测试集合操作"
    
    info "测试 SADD/SISMEMBER"
    set_key="set_test_$(date +%s%N)"
    redis_cmd SADD "$set_key" "member1" "member2" "member3" >/dev/null
    result=$(redis_cmd SISMEMBER "$set_key" "member1")
    [ "$result" = "1" ] || { error "SADD/SISMEMBER测试失败: got '$result'"; return 1; }
    
    info "测试 SREM"
    redis_cmd SREM "$set_key" "member2" >/dev/null
    result=$(redis_cmd SISMEMBER "$set_key" "member2")
    [ "$result" = "0" ] || { error "SREM测试失败: got '$result'"; return 1; }
    
    log "集合操作测试通过"
}

test_hashes() {
    log "测试哈希操作"
    
    info "测试 HSET/HGET"
    hash_key="hash_test_$(date +%s%N)"
    redis_cmd HSET "$hash_key" field1 "value1" field2 "value2" >/dev/null
    result=$(redis_cmd HGET "$hash_key" field1)
    [ "$result" = "value1" ] || { error "HSET/HGET测试失败: got '$result'"; return 1; }
    
    info "测试 HDEL"
    redis_cmd HDEL "$hash_key" field1 >/dev/null
    result=$(redis_cmd HGET "$hash_key" field1)
    [ "$result" = "" ] || { error "HDEL测试失败: got '$result'"; return 1; }
    
    log "哈希操作测试通过"
}

test_sorted_sets() {
    log "测试有序集合操作"
    
    info "测试 ZADD/ZRANGEBYSCORE"
    zset_key="zset_test_$(date +%s%N)"
    redis_cmd ZADD "$zset_key" 1 "member1" 2 "member2" 3 "member3" >/dev/null
    result=$(redis_cmd ZRANGEBYSCORE "$zset_key" 1 2 | tr '\n' ' ')
    [[ "$result" =~ "member1" && "$result" =~ "member2" ]] || { error "ZADD/ZRANGEBYSCORE测试失败: got '$result'"; return 1; }
    
    info "测试 ZREM"
    redis_cmd ZREM "$zset_key" "member2" >/dev/null
    result=$(redis_cmd ZRANGEBYSCORE "$zset_key" 1 3 | tr '\n' ' ')
    [[ "$result" =~ "member1" && "$result" =~ "member3" && ! "$result" =~ "member2" ]] || { error "ZREM测试失败: got '$result'"; return 1; }
    
    log "有序集合操作测试通过"
}

test_expiration() {
    log "测试过期功能"
    
    info "测试 EXPIRE"
    redis_cmd SET expire_key "will_expire" >/dev/null
    redis_cmd EXPIRE expire_key 3 >/dev/null
    result=$(redis_cmd GET expire_key)
    [ "$result" = "will_expire" ] || { error "EXPIRE设置测试失败: got '$result'"; return 1; }
    
    info "等待键过期"
    sleep 4
    result=$(redis_cmd GET expire_key)
    [ "$result" = "" ] || { error "EXPIRE过期测试失败: got '$result'"; return 1; }
    
    log "过期功能测试通过"
}

run_functional_tests() {
    log "开始功能测试"
    
    test_strings || { error "字符串测试失败"; return 1; }
    test_lists || { error "列表测试失败"; return 1; }
    test_sets || { error "集合测试失败"; return 1; }
    test_hashes || { error "哈希测试失败"; return 1; }
    test_sorted_sets || { error "有序集合测试失败"; return 1; }
    test_expiration || { error "过期功能测试失败"; return 1; }
    
    log "所有功能测试通过"
}

run_benchmark() {
    log "开始性能基准测试"
    
    redis_cmd SET benchmark_key "benchmark_value" >/dev/null
    redis_cmd LPUSH benchmark_list "item1" "item2" "item3" >/dev/null
    redis_cmd SADD benchmark_set "member1" "member2" "member3" >/dev/null
    redis_cmd HSET benchmark_hash field1 "value1" field2 "value2" >/dev/null
    redis_cmd ZADD benchmark_zset 1 "member1" 2 "member2" 3 "member3" >/dev/null
    
    info "测试字符串操作性能"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 10000 -t SET,GET --precision 2 2>/dev/null
    
    info "测试批量字符串操作性能"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 1000 -t MSET --precision 2 2>/dev/null
    
    info "测试列表操作性能"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 10000 -t LPUSH,LPOP,RPUSH,RPOP --precision 2 2>/dev/null
    
    info "测试集合操作性能"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 10000 -t SADD,SREM --precision 2 2>/dev/null
    
    info "测试哈希操作性能"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 10000 -t HSET --precision 2 2>/dev/null
    
    info "测试有序集合操作性能"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 10000 -t ZADD --precision 2 2>/dev/null
    
    info "测试混合操作性能"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 100 -n 5000 -t SET,GET,LPUSH,SADD,HSET,ZADD --precision 2 2>/dev/null
    
    info "测试低并发性能 (10连接)"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 10 -n 5000 -t SET,GET --precision 2 2>/dev/null
    
    info "测试高并发性能 (200连接)"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 200 -n 5000 -t SET,GET --precision 2 2>/dev/null
    
    info "测试大数据包性能 (1KB)"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 1000 -t SET,GET -d 1024 --precision 2 2>/dev/null
    
    info "测试小数据包性能 (64字节)"
    redis-benchmark -h $SERVER_HOST -p $SERVER_PORT -c 50 -n 5000 -t SET,GET -d 64 --precision 2 2>/dev/null
    
    info "测试命令延迟分布"
    echo "延迟监控结果 (3秒采样):"
    {
        redis-cli -h $SERVER_HOST -p $SERVER_PORT --latency-history -i 1 &
        LATENCY_PID=$!
        sleep 3
        kill $LATENCY_PID 2>/dev/null || true
        wait $LATENCY_PID 2>/dev/null || true
    } 2>/dev/null
    echo
    
    log "性能基准测试完成"
}

main() {    
    rm -f tests/server.log

    log "开始 TinyRedis 测试"
    
    if ! command -v redis-cli &> /dev/null; then
        error "redis-cli 未安装！请安装Redis客户端工具"
        exit 1
    fi
    
    if ! command -v redis-benchmark &> /dev/null; then
        error "redis-benchmark 未安装！请安装Redis benchmark工具"
        exit 1
    fi
    
    start_server
    
    run_functional_tests
    run_benchmark
    
    log "测试完成"
}

trap 'cleanup; exit 1' INT TERM

main "$@"

{
    cleanup
} >/dev/null 2>&1
