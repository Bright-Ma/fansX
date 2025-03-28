package luaString

type incrBy struct {
	name     string
	function string
}

func (i *incrBy) Name() string {
	return i.name
}

func (i *incrBy) Function() string {
	return i.function
}

var incrByScript *incrBy

func init() {
	incrByScript = &incrBy{}
	incrByScript.name = "string_incrby"
	incrByScript.function = `
local key=KEYS[1]
local num=ARGV[1]

local exists=redis.call("EXISTS",key)
if exists==0
    then return "key not exists"
end

redis.call("INCRBY",key,num)

return "ok"
`
}

func GetIncrBy() *incrBy {
	return incrByScript
}
