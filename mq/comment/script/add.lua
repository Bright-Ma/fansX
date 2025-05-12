redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除
local key=KEYS[1]
local add=ARGV[1]

if redis.call("Exists",key)==0 then
    return
end

redis.call("Incrby",key,add)

return