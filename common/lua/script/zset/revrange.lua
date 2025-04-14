local key=KEYS[1]
local all=ARGV[1]
local b=ARGV[2]
local e=ARGV[3]

local exists=redis.call("EXISTS",key)
if exists==0
    then return nil
end

if all=="true"
    then
    local res=redis.call("ZREVRANGE",key,0,-1)
    return res
end

local res=redis.call("ZREVRANGE",key,b,e)
return res