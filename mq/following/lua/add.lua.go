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

local data=ARGS[1]

local exists=redis.call("EXISTS",key)

if exists==0
    then return
end

for i=0,#data,2
    do
    redis.call("ZADD",tomember(data[i]),data[i+1])
end

return
`
}

func GetAdd() *add { return addScript }
