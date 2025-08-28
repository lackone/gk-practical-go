-- 1, 2, 3, 4, 5, 6, 7
-- ZREMRANGEBYSCORE key 0 6  命令用于移除有序集中，指定分数（score）区间内的所有成员。
-- 执行完后 7

-- 先移除
local key = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 阈值
local rate = tonumber(ARGV[2])
-- 当前时间
local now = tonumber(ARGV[3])
-- 窗口的起始时间
local min = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', min)

local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

if cnt >= rate then
    -- 限流
    return "true"
else
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)

    -- 不限流
    return "false"
end