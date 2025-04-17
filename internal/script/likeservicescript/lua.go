package likeservicescript

import "fansX/internal/middleware/lua"

// List Key:[1]=key,Argv:[2]=timeStamp,[3]=limit
// res:[]interface{} end of res=flag if all record in redis return "true" else return "false"
// res[1*n]=member res[2*n]=score
var List *lua.Script

func init() {
	List = lua.NewScript("userLikeList", `
local key=KEYS[1]
local timeStamp=ARGV[1]
local limit=ARGV[2]

local exists=redis.call("EXISTS",key)

if exists==0
    then return nil
end

local res=redis.call("ZRevRangeByScore",timeStamp,2,"WithScores","Limit",0,limit)
local flag=redis.call("ZScore",key,"all")
if flag[1]==1 then
    table.insert(res,"true")
else
    table.insert(res,"false")
end

return res

`)

}

// BuildList
// Key:[1]=key,[2]=all("true"/"false"),[3]=del("true"/false),[4]=ttl
// Argv: data data[1*n]=score,data[2*n]=member
var BuildList *lua.Script

func init() {
	BuildList = lua.NewScript("buildUserLikeList", `
local key=KEYS[1]
local all=KEYS[2]
local del=KEYS[3]
local ttl=KEYS[4]

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
    redis.call("ZAdd",data[i],data[i+1])
end

if all=="true" then
    redis.call("ZAdd",1,"all")
else
    redis.call("ZAdd",0,"all")
end

return 
`)
}

var Set *lua.Script

func init() {
	Set = lua.NewScript("Set", `
local key=KEYS[1]
local value=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==1 then 
    return 
end 
redis.call("SET",key,value)
return 
`)
}
