-- Atomically removes expired members and renews an existing realtime lease.
-- keys[1]: lease sorted-set key
-- argv[1]: current Unix milliseconds
-- argv[2]: lease expiration Unix milliseconds
-- argv[3]: Redis key TTL in milliseconds
-- argv[4]: lease member identifier
local now = tonumber(ARGV[1])
local expires_at = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])
local member = ARGV[4]

redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now)

if not redis.call('ZSCORE', KEYS[1], member) then
    return 0
end

redis.call('ZADD', KEYS[1], expires_at, member)
redis.call('PEXPIRE', KEYS[1], ttl)

return 1
