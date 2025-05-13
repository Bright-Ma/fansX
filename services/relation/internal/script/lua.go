package script

import "fansX/internal/middleware/lua"

var GetFiled *lua.Script

func init() {
	GetFiled = lua.NewScript("GetFiled", `
local key=KEYS[1]
local field=ARGV[1]
local exists=redis.call("EXISTS",key)
if exists==0
    then return "TableNotExists"
end
local res=redis.call("ZSCORE",key,field)
if not res
then
    return "FieldNotExists"
else
    return tostring(res)
end
`)
}

const GetFiledTableNE = "TableNotExists"
const GetFiledFiledNE = "FieldNotExists"

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

var RevRangeZSet *lua.Script

func init() {
	RevRangeZSet = lua.NewScript("RevRangeZSet", `
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
`)
}
