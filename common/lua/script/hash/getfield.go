package luaHash

type getField struct {
	function string
	name     string
}

func (g *getField) Name() string     { return g.name }
func (g *getField) Function() string { return g.function }

var getFieldScript *getField

func init() {
	getFieldScript = &getField{}
	getFieldScript.name = "hash_getfield"
	getFieldScript.function = `
local name=KEYS[1]
local field=ARGS[1]
local exists=redis.call("EXISTS",name)
if exists==0 then
    return "TableNotExists"
end
local res=redis.call("HGet",name,field)
if res==nil
then
    return "FieldNotExists"
else
    return res
end`
}

func GetGetField() *getField {
	return getFieldScript
}

var HGetFieldNotExists = "FieldNotExists"
var HGetTableNotExists = "TableNotExists"
