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
    redis.call("HSet",key,data[i],data[i+1])
end
redis.call("EXPIRE",key,tonumber(ttl))

return true