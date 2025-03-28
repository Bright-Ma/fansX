local key=KEYS[1]
local field=ARGV[1]
local exists=redis.call("EXISTS",key)
if exists==0
    then return "TableNotExists"
end
local res=redis.call("ZSCORE",key,field)
if not res
then
    return "FieldNotExists"
else
    return tostring(res)
end