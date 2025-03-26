package luaHash

import (
	"os"
	"runtime"
)

type create struct {
	function string
}

func (c *create) Name() string     { return "hash_create" }
func (c *create) Function() string { return c.function }

var createScript *create

func init() {
	_, path, _, _ := runtime.Caller(0)
	res, err := os.ReadFile(path + "create.lua")
	if err != nil {
		panic(err.Error())
	}
	createScript = &create{string(res)}
}

func GetCreate() *create {
	return createScript
}
