local val = redis.call('get', KEYS[1])
local interval = ARGV[1]
local limit = tonumber(ARGV[2])
if val == false then
    if limit < 1 then
        --限流
        return "true"
    else
        -- key 不存在，设置初始值1
        redis.call('set', KEYS[1], 1, 'PX', interval)
        -- 不执行限流
        return "false"
    end
elseif tonumber(val) < limit then
    -- 自增1
    redis.call('incr', KEYS[1])
    -- 不执行限流
    return "false"
else
    --限流
    return "true"
end