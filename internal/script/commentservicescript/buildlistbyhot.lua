redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local timeout=KEYS[2]
local ttl=KEYS[3]
local data=ARGV

local exists=redis.call("EXISTS",key)
if exists==1 then
    redis.call("ZRange",key,"-inf",)
end

for i=1,#data,2
do redis.call("ZAdd",key,data[i],data[i+1])
end

return