redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除

local key=KEYS[1]
local value=ARGV[1]
local ttl=ARGV[2]

local res=redis.call("Get",key)
if res==nil then
return nil
end

if res~=value then
return nil
end

redis.call("Expire",key,ttl)

return 1

