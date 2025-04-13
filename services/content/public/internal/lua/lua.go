package interlua

type revRangeByScoreWithScores struct {
	name     string
	function string
}

func (r *revRangeByScoreWithScores) Name() string {
	return r.name
}
func (r *revRangeByScoreWithScores) Function() string { return r.function }

var revScript *revRangeByScoreWithScores

func init() {
	revScript = &revRangeByScoreWithScores{}
	revScript.name = "revRangeByScoreWithScores"
	revScript.function = `
local key=KEYS[1]
local max=ARGV[1]
local min=ARGV[2]
local exists=redis.call("EXISTS",key)
if exists==0
    then return nil
end
local res=redis.call("ZREVRANGEBYSCORE",key,max,min,"WITHSCORES")

return res

`
}

func GetRevScript() *revRangeByScoreWithScores {
	return revScript
}
