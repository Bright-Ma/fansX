package commentservicescript

import "fansX/internal/middleware/lua"

var GetCountScript *lua.Script

func init() {
	GetCountScript = lua.NewScript("get_count", `
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

`)
}

var GetByHot *lua.Script

func init() {
	GetByHot = lua.NewScript("get_by_hot", `
local key=KEYS[1]
local limit=ARGV[1]
local offset=ARGV[2]


local exists=redis.call("EXISTS",key)
if exists==0 then
    return {3}
end

local status=0

local ex=redis.call("ZRange",key,0,0)
local ttl=redis.call("TTL",key)
if tonumber(ex[1])>=tonumber(ttl) then
    status=2
else
    status=1
end

local res=redis.call("ZRevRangeByScore",key,"+inf",0,"Limit",tonumber(offset),tonumber(limit))
table.insert(res,status)
return res
`)
}

var Build *lua.Script

func init() {
	Build = lua.NewScript("build", `
local key=KEYS[1]
local ttl=KEYS[2]
local data=ARGV

local exists=redis.call("EXISTS",key)
if exists==1 then
    local ex=redis.call("ZRange",key,0,0)
    local last=redis.call("TTL",key)
    if tonumber(ex[1])>=tonumber(last) then
        redis.call("DEL",key)
    else
        return
    end
end

for i=1,#data,2
do redis.call("ZAdd",key,data[i],data[i+1])
end

redis.call("EXPIRE",key,ttl)

return
`)
}

var GetByTime *lua.Script

func init() {
	GetByTime = lua.NewScript("get_by_time", `
local key=KEYS[1]
local limit=ARGV[1]
local timestamp=ARGV[2]


local exists=redis.call("EXISTS",key)
if exists==0 then
    return {3}
end

local status=0

local ex=redis.call("ZRange",key,0,0)
local ttl=redis.call("TTL",key)
if tonumber(ex[1])>=tonumber(ttl) then
    status=2
else
    status=1
end

local res=redis.call("ZRevRangeByScore",key,timestamp,0,"Limit",0,tonumber(limit))
table.insert(res,status)
return res

`)
}
