-- Atomically removes expired members and acquires or renews one realtime lease.
-- keys[1]: lease sorted-set key
-- argv[1]: current Unix milliseconds
-- argv[2]: lease expiration Unix milliseconds
-- argv[3]: maximum active members
-- argv[4]: Redis key TTL in milliseconds
-- argv[5]: lease member identifier
local now = tonumber(ARGV[1])
local expires_at = tonumber(ARGV[2])
local maximum_members = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])
local member = ARGV[5]

redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now)

if redis.call('ZSCORE', KEYS[1], member) then
    redis.call('ZADD', KEYS[1], expires_at, member)
    redis.call('PEXPIRE', KEYS[1], ttl)

    return { 1, redis.call('ZCARD', KEYS[1]) }
end

if redis.call('ZCARD', KEYS[1]) >= maximum_members then
    return { 0, redis.call('ZCARD', KEYS[1]) }
end

redis.call('ZADD', KEYS[1], expires_at, member)
redis.call('PEXPIRE', KEYS[1], ttl)

return { 1, redis.call('ZCARD', KEYS[1]) }
