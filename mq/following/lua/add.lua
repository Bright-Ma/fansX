local key=KEYS[1]

local data=ARGS[1]

local exists=redis.call("EXISTS",key)

if exists==0
    then return
end

for i=0,#data,2
    do
    redis.call("ZADD",tomember(data[i]),data[i+1])
end

return

