package luaHash

type create struct {
	name     string
	function string
}

func (c *create) Name() string     { return c.name }
func (c *create) Function() string { return c.function }

var createScript *create

func init() {
	createScript = &create{}
	createScript.name = "hash_create"
	createScript.function = `
local key=KEYS[1]
local del=KEYS[2]
local data=ARGS

if (#data)%2~=0
    then return {err="data nums should be 2*x"}
end

exists=redis.call("EXISTS",key)
if exists==1
    then
    if del=="true"
        then redis.call("DEL",key)
        else return
    end
end

for i=1,#data,2
do
    redis.call("HSet",key,data[i],data[i+1])
end

return
`
}

func GetCreate() *create {
	return createScript
}
