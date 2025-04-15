local zset_key = KEYS[1]
local target_prefix = ARGV[1] .. ';'
local cursor = "0"
local total_deleted = 0

repeat
    local result = redis.call('ZSCAN', zset_key, cursor, 'MATCH', target_prefix .. '*', 'COUNT', 100)
    cursor = result[1]
    local elements = result[2]

    for i = 1, #elements, 2 do  -- ZSCAN返回 [member1, score1, member2, score2, ...]
        redis.call('ZREM', zset_key, elements[i])
        total_deleted = total_deleted + 1
    end
until cursor == "0"

return total_deleted

