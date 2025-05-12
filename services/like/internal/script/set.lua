redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local value=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==1 then
    return
end
redis.call("SET",key,value)
return