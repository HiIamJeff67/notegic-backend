#!lua name=rate_limit_record_library

-- batch_synchronize_rate_limit_record_by_formatted_keys:
-- Redis functions to batch synchronize the rate limit record by the given formatted keys
--
-- keys: array of formatted keys
-- argv: array of json objects containing synchronizeDto
-- Format of argv: [num_of_changing_tokens_1, is_accumulated_1, num_of_changing_tokens_2, is_accumulated_2, ...]
--                 each `NUM_OF_ARGV_PER_KEY` mapping to a key in `keys`
local BATCH_SYNC_NUM_OF_ARGV_PER_KEY = 2
local function batch_synchronize_rate_limit_record_by_formatted_keys(keys, argv)
    if #keys == 0 or #argv == 0 then
        return { updated_count = 0, error = nil }
    end

    if #argv % BATCH_SYNC_NUM_OF_ARGV_PER_KEY ~= 0 or #keys * BATCH_SYNC_NUM_OF_ARGV_PER_KEY ~= #argv then
        return { updated_count = 0, error = "The size or the format of the argument variables don't match the keys" }
    end

    local updated_count = 0
    for key_index = 1, #keys do
        local argv_index = (key_index - 1) * BATCH_SYNC_NUM_OF_ARGV_PER_KEY + 1

        local key = keys[key_index]
        local is_accumulated = argv[argv_index] == 'true'
        local num_of_changing_tokens = tonumber(argv[argv_index + 1]) or 0

        local cache_string = redis.call('GET', key)

        if cache_string then
            local is_decode_ok, cache = pcall(function()
                return cjson.decode(cache_string)
            end)

            if is_decode_ok and cache.numOfTokens >= 0 then
                cache.numOfTokens = is_accumulated
                    and (cache.numOfTokens + num_of_changing_tokens)
                    or math.max(0, cache.numOfTokens - num_of_changing_tokens)

                cache.updatedAt = redis.call('TIME')[1]

                local is_encode_ok, json_str = pcall(function()
                    return cjson.encode(cache)
                end)

                if is_encode_ok then
                    -- get the TTL first
                    local ttl = redis.call('TTL', key)
                    -- then update the value with modifying the expiration time
                    redis.call('SET', key, json_str)

                    -- if there's still some time to expiry, then restore the expiration time of it
                    -- note that if we don't set the expiration time back, it will expiry SOON
                    if ttl > 0 then
                        redis.call('EXPIRE', key, ttl)
                    end

                    updated_count = updated_count + 1
                end
            end
        end
    end

    return { updated_count = updated_count, error = nil }
end

-- batch_delete_rate_limit_record_by_formatted_keys:
-- Redis functions to batch delete the rate limit record by the given formatted keys
--
-- keys: array of formatted keys
-- argv: a placeholder for argv, but we don't use it here
local function batch_delete_rate_limit_record_by_formatted_keys(keys, _)
    if #keys == 0 then
        return { deleted_count = 0, error = nil }
    end

    local deleted_count = 0
    for key_index = 1, #keys do
        local key = keys[key_index]

        local result = redis.call('DEL', key)
        if result == 1 then
            deleted_count = deleted_count + 1
        end
    end

    return { deleted_count = deleted_count, error = nil }
end

redis.register_function('batch_synchronize_rate_limit_record_by_formatted_keys',
    batch_synchronize_rate_limit_record_by_formatted_keys)
redis.register_function('batch_delete_rate_limit_record_by_formatted_keys',
    batch_delete_rate_limit_record_by_formatted_keys)
