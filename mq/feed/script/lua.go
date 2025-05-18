package script

import "fansX/internal/middleware/lua"

var ZSetAdd *lua.Script

func init() {
	ZSetAdd = lua.NewScript("ZSetAdd", `
local key=KEYS[1]
local ma=KEYS[2]
local data=ARGV

local exists=redis.call("EXISTS",key)
if exists==0
    then return true
end

for i=1,#data,2
   do
    redis.call("ZADD",key,tonumber(data[i]),data[i+1])
end

local nums=redis.call("ZCARD",key)
if tonumber(ma)<tonumber(nums)
    then
    redis.call("ZREMRANGEBYRANK",key,0,tonumber(nums)-tonumber(ma)-1)
end

return true
`)
}
