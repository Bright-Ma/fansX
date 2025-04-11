local key=KEYS[1]

local data=ARGV

local exists=redis.call("EXISTS",key)

if exists==0
    then return nil
end

for i=1,#data,2
    do
    redis.call("ZADD",key,tonumber(data[i]),data[i+1])
end

return true

