redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local limit=ARGV[1]
local timestamp=ARGV[2]


local exists=redis.call("EXISTS",key)
if exists==0 then
    return {3}
end

local status=0

local ex=redis.call("ZRange",key,0,0)
local ttl=redis.call("TTL",key)
if tonumber(ex[1])>=tonumber(ttl) then
    status=2
else
    status=1
end

local res=redis.call("ZRevRangeByScore",key,timestamp,0,"Limit",0,tonumber(limit))
table.insert(res,status)
return res
