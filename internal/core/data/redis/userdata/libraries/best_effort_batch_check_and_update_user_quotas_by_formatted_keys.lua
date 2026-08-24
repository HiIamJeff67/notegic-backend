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

local function best_effort_batch_check_and_update_user_quotas_by_formatted_keys(keys, argv)
    if #keys == 0 or #argv == 0 then return { updated_count = 0, error = nil } end
    if #argv % 4 ~= 0 then return { updated_count = 0, error = "Argv size mismatch" } end

    local updated_count = 0
    for index = 1, #keys do
        local offset = (index - 1) * 4
        local cache_string = redis.call('GET', keys[index])
        if cache_string then
            local cache = safe_decode(cache_string)
            if cache then
                local field = argv[offset + 1]
                local change_amount = tonumber(argv[offset + 2]) or 0
                local max_limit = tonumber(argv[offset + 3]) or 0
                local ttl = tonumber(argv[offset + 4]) or 0
                local new_value = (tonumber(cache[field]) or 0) + change_amount
                local valid = (change_amount > 0 and new_value <= max_limit) or
                    (change_amount <= 0 and new_value >= 0)
                if valid then
                    cache[field] = new_value
                    local new_json = safe_encode(cache)
                    if new_json then
                        redis.call('SET', keys[index], new_json)
                        redis.call('EXPIRE', keys[index], ttl)
                        updated_count = updated_count + 1
                    end
                end
            end
        end
    end

    return { updated_count = updated_count, error = nil }
end

redis.register_function(
    'best_effort_batch_check_and_update_user_quotas_by_formatted_keys',
    best_effort_batch_check_and_update_user_quotas_by_formatted_keys
)
