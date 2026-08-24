-- Removes one realtime lease member and deletes its key when it is empty.
-- keys[1]: lease sorted-set key
-- argv[1]: lease member identifier
redis.call('ZREM', KEYS[1], ARGV[1])

if redis.call('ZCARD', KEYS[1]) == 0 then
    redis.call('DEL', KEYS[1])
end

return 1
