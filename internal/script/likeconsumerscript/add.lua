redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local add=KEYS[2]
local nums=ARGV[1]

local exists=redis.call("exists",key)

if exists==0 and add=="false" then
    return
end

redis.call("IncrBy",key,nums)

return