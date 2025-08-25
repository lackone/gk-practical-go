if redis.call('get', KEYS[1]) == ARGV[1] then
    -- 是你的锁
    return redis.call('EXPIRE', KEYS[1], ARGV[2])
else
    -- 不是你的锁
    return 0
end