local key=KEYS[1]
local num=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
    then return "key not exists"
end

redis.call("INCRBY",key,num)

return "ok"