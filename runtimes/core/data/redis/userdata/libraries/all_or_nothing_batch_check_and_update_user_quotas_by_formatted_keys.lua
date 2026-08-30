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

local function all_or_nothing_batch_check_and_update_user_quotas_by_formatted_keys(keys, argv)
    if #keys == 0 or #argv == 0 then return { updated_count = 0, error = nil } end
    if #argv ~= #keys * 4 then return { updated_count = 0, error = "Argv size mismatch" } end

    local pending_updates = {}
    for index = 1, #keys do
        local offset = (index - 1) * 4
        local cache_string = redis.call('GET', keys[index])
        if not cache_string then return { updated_count = 0, error = "Cache not found" } end

        local cache = safe_decode(cache_string)
        if not cache then return { updated_count = 0, error = "JSON decode failed" } end

        local field = argv[offset + 1]
        local change_amount = tonumber(argv[offset + 2]) or 0
        local max_limit = tonumber(argv[offset + 3]) or 0
        local ttl = tonumber(argv[offset + 4]) or 0
        local new_value = (tonumber(cache[field]) or 0) + change_amount
        if change_amount > 0 and new_value > max_limit then
            return { updated_count = 0, error = "Quota exceeded" }
        end
        if change_amount <= 0 and new_value < 0 then
            return { updated_count = 0, error = "Negative quota" }
        end

        cache[field] = new_value
        table.insert(pending_updates, { key = keys[index], cache = cache, ttl = ttl })
    end

    for _, update in ipairs(pending_updates) do
        local new_json = safe_encode(update.cache)
        if not new_json then return { updated_count = 0, error = "JSON encode failed" } end
        redis.call('SET', update.key, new_json)
        redis.call('EXPIRE', update.key, update.ttl)
    end

    return { updated_count = #keys, error = nil }
end

redis.register_function(
    'all_or_nothing_batch_check_and_update_user_quotas_by_formatted_keys',
    all_or_nothing_batch_check_and_update_user_quotas_by_formatted_keys
)
