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

local function check_and_update_user_quota_by_formatted_key(keys, argv)
    if #keys ~= 1 or #argv ~= 4 then
        return { new_value = -1, error = "Invalid arguments" }
    end

    local key = keys[1]
    local field = argv[1]
    local change_amount = tonumber(argv[2]) or 0
    local max_limit = tonumber(argv[3]) or 0
    local ttl = tonumber(argv[4]) or 0
    local cache_string = redis.call('GET', key)
    if not cache_string then return { new_value = -1, error = "Cache not found" } end

    local cache = safe_decode(cache_string)
    if not cache then return { new_value = -1, error = "Failed to decode JSON" } end

    local current = tonumber(cache[field]) or 0
    local new_value = current + change_amount
    if change_amount > 0 and new_value > max_limit then
        return { new_value = -1, error = "Quota exceeded" }
    end
    if change_amount <= 0 and new_value < 0 then
        return { new_value = -1, error = "Quota cannot be negative" }
    end

    cache[field] = new_value
    local new_json = safe_encode(cache)
    if not new_json then return { new_value = -1, error = "Failed to encode JSON" } end

    redis.call('SET', key, new_json)
    redis.call('EXPIRE', key, ttl)
    return { new_value = new_value, error = nil }
end

redis.register_function('check_and_update_user_quota_by_formatted_key', check_and_update_user_quota_by_formatted_key)
