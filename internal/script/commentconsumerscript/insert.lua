redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local member=ARGV[1]
local score=ARGV[2]

if redis.call("Exists",key)==0 then
    return
end

local ex=redis.call("ZRange",key,0,1)
local ttl=redis.call("TTL",key)
if tonumber(ex[2])>=tonumber(ttl) then
    return
end

redis.call("ZAdd",key,score,member)
return 

