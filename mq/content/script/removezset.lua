redis =require"redis"
local KEYS={}
local ARGV={}
--上述三行为了防止报错，在使用时将其删除local key=KEYS[1]

local key=KEYS[1]
local member=ARGV[1]
if redis.call("EXISTST",key)==0 then
return true
end

return true
