redis =require"redis"
local KEYS={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]

local exists=redis.call("EXISTS",key)
if exists==0 then
return {0,3}
end

local str=redis.call("GET",key)
local part1, part2 = string.match(str, "^(.-);(.*)$")
local res={part1}

if tonumber(part2)>=tonumber(redis.call("TTL",key)) then
    table.insert(res,2)
else
    table.insert(res,1)
end

return res