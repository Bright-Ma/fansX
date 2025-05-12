redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除
local num=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
    then return "key not exists"
end

redis.call("INCRBY",key,num)

return "ok"