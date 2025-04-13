local key=KEYS[1]
local max=ARGV[1]
local min=ARGV[2]
local exists=redis.call("EXISTS",key)
if exists==0
    then return nil
end
local res=redis.call("ZREVRANGEBYSCORE",key,max,min,"WITHSCORES")

return res

