package commentconsumerscript

import "fansX/internal/middleware/lua"

var Insert *lua.Script

func init() {
	Insert = lua.NewScript("insert", `
local key=KEYS[1]
local member=ARGV[1]
local score=ARGV[2]

if redis.call("Exists",key)==0 then
    return
end

local ex=redis.call("ZRange",key,0,1)
local ttl=redis.call("TTL",key)
if tonumber(ex[2])>=tonumber(ttl) then
    return
end

redis.call("ZAdd",key,score,member)
return
`)
}

var Add *lua.Script

func init() {
	Add = lua.NewScript("add", `
local key=KEYS[1]
local add=ARGV[1]

if redis.call("Exists",key)==0 then
    return
end

redis.call("Incrby",key,add)

return 
`)
}
