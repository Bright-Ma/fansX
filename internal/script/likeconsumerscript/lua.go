package likeconsumerscript

import "fansX/internal/middleware/lua"

var InsertScript *lua.Script

func init() {
	InsertScript = lua.NewScript("insert", `
local key=KEYS[1]
local del=KEYS[2]

local data=ARGV

local exists=redis.call("EXISTS",key)

if exists==1 then
    if del=="true" then
        redis.call("DEL",key)
    else
        return
    end
end

for i=1,#data,2
do
    redis.call("ZAdd",key,data[i],data[i+1])
end

local res=redis.call("ZCard",key)

if tonumber(res)<=1001 then
return
end

redis.call("ZAdd",key,0,"all")

return
`)
}

var AddScript *lua.Script

func init() {
	AddScript = lua.NewScript("add", `
local key=KEYS[1]
local add=KEYS[2]
local nums=ARGV[1]

local exists=redis.call("exists",key)

if exists==0 and add=="false" then
    return
end

redis.call("IncrBy",key,nums)

return 
`)
}
