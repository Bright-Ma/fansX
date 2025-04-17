local key=KEYS[1]
local value=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==1 then
    return
end
redis.call("SET",key,value)
return