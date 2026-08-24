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

local function all_or_nothing_batch_check_and_update_user_quotas_by_formatted_key(keys, argv)
    if #keys ~= 1 or #argv == 0 or #argv % 4 ~= 0 then
        return { updated_count = 0, error = "Invalid arguments" }
    end

    local cache_string = redis.call('GET', keys[1])
    if not cache_string then return { updated_count = 0, error = "Cache not found" } end

    local cache = safe_decode(cache_string)
    if not cache then return { updated_count = 0, error = "JSON decode failed" } end

    local pending_values = {}
    local max_ttl = 0
    for index = 0, (#argv / 4) - 1 do
        local offset = index * 4
        local field = argv[offset + 1]
        local change_amount = tonumber(argv[offset + 2]) or 0
        local max_limit = tonumber(argv[offset + 3]) or 0
        local ttl = tonumber(argv[offset + 4]) or 0
        local current = pending_values[field] or (tonumber(cache[field]) or 0)
        local new_value = current + change_amount
        if change_amount > 0 and new_value > max_limit then
            return { updated_count = 0, error = "Quota exceeded" }
        end
        if change_amount <= 0 and new_value < 0 then
            return { updated_count = 0, error = "Negative quota" }
        end

        pending_values[field] = new_value
        if ttl > max_ttl then max_ttl = ttl end
    end

    for field, value in pairs(pending_values) do
        cache[field] = value
    end

    local new_json = safe_encode(cache)
    if not new_json then return { updated_count = 0, error = "JSON encode failed" } end
    redis.call('SET', keys[1], new_json)
    redis.call('EXPIRE', keys[1], max_ttl)

    return { updated_count = #argv / 4, error = nil }
end

redis.register_function(
    'all_or_nothing_batch_check_and_update_user_quotas_by_formatted_key',
    all_or_nothing_batch_check_and_update_user_quotas_by_formatted_key
)
