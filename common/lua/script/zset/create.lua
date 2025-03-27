local key=KEYS[1]
local del=KEYS[2]
local data=ARGS[1]

if (#data)%2~=0
then return {err="data nums should be 2*x"}
end

local exists=redis.call("EXISTS",key)

if exists==1
    then
    if del=="true"
        then redis.call("DEL",key)
        else return
    end
end

for i=1,#data,2
    do
    local score=tonumber(data[i])
    local value=data[i+1]
    redis.call("ZADD",score,value)
end
return