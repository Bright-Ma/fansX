local key=KEYS[1]
local limit=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
then return nil
end

local res=redis.call("ZREVRANGE",0,tonumber(limit),"WITHSCORES")
return res