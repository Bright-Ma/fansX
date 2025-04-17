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
