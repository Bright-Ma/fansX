package luaFollowing

type add struct {
	function string
	name     string
}

func (a *add) Name() string {
	return a.name
}
func (a *add) Function() string { return a.function }

var addScript *add

func init() {
	addScript = &add{}
	addScript.name = "add"
	addScript.function = `
local key=KEYS[1]

local data=ARGV

local exists=redis.call("EXISTS",key)

if exists==0
    then return nil
end

for i=1,#data,2
    do
    redis.call("ZADD",key,tonumber(data[i]),data[i+1])
end

return true
`
}

func GetAdd() *add { return addScript }
