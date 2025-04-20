redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local add=KEYS[2]

local data=ARGV

local exists=redis.call("EXISTS",key)

if exists==0 and add=="false" then
    return
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


