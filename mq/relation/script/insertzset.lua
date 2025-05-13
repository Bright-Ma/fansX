redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除
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
