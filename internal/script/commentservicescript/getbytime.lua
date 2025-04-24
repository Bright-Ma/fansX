redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除
-- -1 extra[2] cache timeout -2 extra[1] is all
local 	StatusFind        = 1 << 1
local	StatusNeedRebuild = 1 << 2
local 	StatusNotFind     = 1 << 3
local	StatusIsAll       = 1 << 4
local	StatusNotAll      = 1 << 5

local key=KEYS[1]
local limit=ARGV[1]
local timestamp=ARGV[2]


local exists=redis.call("EXISTS",key)
if exists==0 then
    return {StatusNotFind}
end

local status=0

local extra=redis.call("ZRange",key,0,1)
local ttl=redis.call("TTL",key)
if tonumber(extra[2])>=tonumber(ttl) then
    status=StatusNeedRebuild
else
    status=StatusFind
end
if extra[1]=="all" then
    status=status|StatusIsAll
else
    status=status|StatusNotAll
end

local res=redis.call("ZRevRangeByScore",key,timestamp,0,"Limit",0,tonumber(limit))
table.insert(res,status)
return res
