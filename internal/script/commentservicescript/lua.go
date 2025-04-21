package commentservicescript

import "fansX/internal/middleware/lua"

var GetCountScript *lua.Script

func init() {
	GetCountScript = lua.NewScript("get_count", `
local key=KEYS[1]

local exists=redis.call("EXISTS",key)
if exists==0 then
return nil
end

local res=redis.call("GET",key)

return res
`)
}
