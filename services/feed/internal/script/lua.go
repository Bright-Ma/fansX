package script

import "fansX/internal/middleware/lua"

var RevRange *lua.Script

func init() {
	RevRange = lua.NewScript("RevRange", `
local key=KEYS[1]
local limit=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
then return nil
end

local res=redis.call("ZREVRANGE",0,tonumber(limit),"WITHSCORES")
return res
`)
}

var RangeByScore *lua.Script

func init() {
	RangeByScore = lua.NewScript("RangeByScore", `
local key=KEYS[1]
local limit=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
then return nil
end

local res=redis.call("ZREVRANGE",0,tonumber(limit),"WITHSCORES")
return res
`)
}

var BuildZSet *lua.Script

func init() {
	BuildZSet = lua.NewScript("BuildZSet", `
local key=KEYS[1]
local del=KEYS[2]
local ttl=KEYS[3]
local data=ARGV

if (#data)%2~=0
then return {err="data nums should be 2*x"}
end

local exists=redis.call("EXISTS",key)

if exists==1
    then
    if del=="true"
        then redis.call("DEL",key)
        else return true
    end
end

for i=1,#data,2
    do
    local score=tonumber(data[i])
    local value=data[i+1]
    redis.call("ZADD",key,score,value)
end
redis.call("EXPIRE",key,tonumber(ttl))

return true
`)

}
