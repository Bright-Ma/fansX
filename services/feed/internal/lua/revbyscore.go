package interlua

type revByScore struct {
	name     string
	function string
}

func (r *revByScore) Name() string {
	return r.name
}
func (r *revByScore) Function() string { return r.function }

var revByScoreScript *revByScore

func init() {
	revByScoreScript = &revByScore{}
	revByScoreScript.name = "revByScore"
	revByScoreScript.function = `
local key=KEYS[1]
local min=ARGV[1]
local max=ARGV[2]

local exists=redis.call("EXISTS",key)
if exists==0
    then return nil
end

local res=redis.call("ZRANGEBYSCORE",key,tonumber(min),tonumber(max),"WITHSCORES")

return res
`
}

func GetRevByScoreScript() *revByScore {
	return revByScoreScript
}
