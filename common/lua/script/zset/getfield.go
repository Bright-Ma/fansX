package luaZset

type getField struct {
	function string
	name     string
}

func (g *getField) Name() string     { return g.name }
func (g *getField) Function() string { return g.function }

var getFieldScript *getField

func init() {
	getFieldScript = &getField{}
	getFieldScript.name = "zset.getfield"
	getFieldScript.function = `
local key=KEYS[1]
local field=ARGS[1]
local exists=redis.call("EXISTS",name)
if exists==0
    then return "TableNotExists"
end
local res=redis.call("ZSCORE",key,field)
if res==nil
then
    return "FieldNotExists"
else
    return tostring(res)
end
`
}

func GetGetField() *getField {
	return getFieldScript
}

var FieldNotExists = "FieldNotExists"
var TableNotExists = "TableNotExists"
