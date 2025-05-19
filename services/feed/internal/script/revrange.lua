redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local limit=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
then return nil
end

local res=redis.call("ZREVRANGE",0,tonumber(limit),"WITHSCORES")
return res