package luaZset

import (
	"os"
	"runtime"
)

type revRange struct {
	function string
}

func (r *revRange) Name() string {
	return "zset_revrange"
}
func (r *revRange) Function() string { return r.function }

var revRangeScript *revRange

func init() {
	_, path, _, _ := runtime.Caller(0)
	res, err := os.ReadFile(path + "revrange.lua")
	if err != nil {
		panic(err.Error())
	}
	revRangeScript = &revRange{string(res)}
}

// GetRevRange
// input KEYS[1]=key ARGS[1]=start ARGS[2]=end
// return nil if key not exists
func GetRevRange() *revRange { return revRangeScript }
