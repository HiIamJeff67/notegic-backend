local function safe_decode(value)
    local ok, result = pcall(cjson.decode, value)
    if ok then return result end
    return nil
end

local function safe_encode(value)
    local ok, result = pcall(cjson.encode, value)
    if ok then return result end
    return nil
end

local function best_effort_batch_check_and_update_user_quotas_by_formatted_key(keys, argv)
    if #keys ~= 1 or #argv == 0 or #argv % 4 ~= 0 then
        return { updated_count = 0, error = "Invalid arguments" }
    end

    local cache_string = redis.call('GET', keys[1])
    if not cache_string then return { updated_count = 0, error = "Cache not found" } end

    local cache = safe_decode(cache_string)
    if not cache then return { updated_count = 0, error = "JSON decode failed" } end

    local updated_count = 0
    local max_ttl = 0
    for index = 0, (#argv / 4) - 1 do
        local offset = index * 4
        local field = argv[offset + 1]
        local change_amount = tonumber(argv[offset + 2]) or 0
        local max_limit = tonumber(argv[offset + 3]) or 0
        local ttl = tonumber(argv[offset + 4]) or 0
        local new_value = (tonumber(cache[field]) or 0) + change_amount
        local valid = (change_amount > 0 and new_value <= max_limit) or
            (change_amount <= 0 and new_value >= 0)
        if valid then
            cache[field] = new_value
            updated_count = updated_count + 1
            if ttl > max_ttl then max_ttl = ttl end
        end
    end

    if updated_count > 0 then
        local new_json = safe_encode(cache)
        if not new_json then return { updated_count = 0, error = "JSON encode failed" } end
        redis.call('SET', keys[1], new_json)
        redis.call('EXPIRE', keys[1], max_ttl)
    end

    return { updated_count = updated_count, error = nil }
end

redis.register_function(
    'best_effort_batch_check_and_update_user_quotas_by_formatted_key',
    best_effort_batch_check_and_update_user_quotas_by_formatted_key
)
