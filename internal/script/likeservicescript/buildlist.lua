redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

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

redis.call("EXPIRE",key,ttl)

return