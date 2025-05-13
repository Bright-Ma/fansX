package script

import "fansX/internal/middleware/lua"

var InsertZSet *lua.Script

func init() {
	InsertZSet = lua.NewScript("InsertZSet", `
local key=KEYS[1]
local data=ARGV

local exists=redis.call("EXISTS",key)

if exists==0
    then return true
end

for i=1,#data,2
    do
    redis.call("ZADD",key,tonumber(data[i]),data[i+1])
end

return true

`)
}

var IncrBy *lua.Script

func init() {
	IncrBy = lua.NewScript("IncrBy", `
local key=KEYS[1]
local num=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
    then return "key not exists"
end

redis.call("INCRBY",key,num)

return "ok"
`)
}

var InsertZSetWithMa *lua.Script

func init() {
	InsertZSetWithMa = lua.NewScript("InsertZSetWithMa", `
local key=KEYS[1]
local ma=KEYS[2]
local data=ARGV

local exists=redis.call("EXISTS",key)
if exists==0
    then return true
end

for i=1,#data,2
   do
    redis.call("ZADD",key,tonumber(data[i]),data[i+1])
end

local nums=redis.call("ZCARD",key)
if tonumber(ma)<tonumber(nums)
    then
    redis.call("ZREMRANGEBYRANK",key,0,tonumber(nums)-tonumber(ma)-1)
end

return true

`)
}

var RemoveZSet *lua.Script

func init() {
	RemoveZSet = lua.NewScript("RemoveZSet", `
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
`)
}
