package luaZset

type revRange struct {
	function string
	name     string
}

func (r *revRange) Name() string {
	return r.name
}
func (r *revRange) Function() string { return r.function }

var revRangeScript *revRange

func init() {
	revRangeScript = &revRange{}
	revRangeScript.name = "zset_revrange"
	revRangeScript.function = `
local key=KEYS[1]
local all=ARGS[2]
local b=ARGS[2]
local e=ARGS[3]

local exists=redis.call("EXISTS",key)
if exists==0
    then return nil
end

if all=="true"
    then
    local res=redis.call("ZREVRANGE",key,0,-1)
    return res
end

local res=redis.call("ZREVRANGE",key,b,e)
return res
`
}

// GetRevRange
// input KEYS[1]=key ARGS[1]=start ARGS[2]=end
// return nil if key not exists
func GetRevRange() *revRange { return revRangeScript }
