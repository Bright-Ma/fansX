package interlua

type revRange struct {
	name     string
	function string
}

func (r *revRange) Name() string {
	return r.name
}
func (r *revRange) Function() string { return r.function }

var revRangeScript *revRange

func init() {
	revRangeScript = &revRange{}
	revRangeScript.name = "revRange"
	revRangeScript.function = `
local key=KEYS[1]
local limit=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
then return nil
end

local res=redis.call("ZREVRANGE",0,tonumber(limit),"WITHSCORES")
return res
`
}

func GetRevRangeScript() *revRange {
	return revRangeScript
}
