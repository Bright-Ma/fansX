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

var GetCommentListByHot *lua.Script

func init() {
	GetCommentListByHot = lua.NewScript("get_comment_list_by_hot", `
local key=KEYS[1]
local limit=ARGV[1]
local offset=ARGV[2]

local exists=redis.call("EXISTS",key)
if exists==0 then
return nil
end

local res=redis.call("ZRevRangeByScore","+inf","-inf","WithScores","Limit",tonumber(offset),tonumber(limit))

return res
`)
}
