local name=KEYS[1]
local field=ARGS[1]
local exists=redis.call("EXISTS",name)
if exists==0 then
    return "TableNotExists"
end
local res=redis.call("HGet",name,field)
if res==nil
then
    return "FieldNotExists"
else
    return res
end