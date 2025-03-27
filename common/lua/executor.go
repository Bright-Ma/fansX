package lua

import (
	luaHash "bilibili/common/lua/script/hash"
	luaString "bilibili/common/lua/script/string"
	luaZset "bilibili/common/lua/script/zset"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type Executor struct {
	client *redis.Client
	sha    map[string]string
}

func NewExecutor(client *redis.Client) *Executor {
	return &Executor{
		sha:    make(map[string]string),
		client: client,
	}
}

func (e *Executor) Load(ctx context.Context, scripts []Script) (int, error) {
	for i, script := range scripts {
		_, ok := e.sha[script.Name()]
		if ok {
			return i + 1, errors.New("repeat script name:" + fmt.Sprint(script.Name()))
		}
		res, err := e.client.ScriptLoad(ctx, script.Function()).Result()
		if err != nil {
			return i + 1, err
		}
		e.sha[script.Name()] = res
	}

	fmt.Print("Load:")
	for i, script := range scripts {
		fmt.Print(script.Name())
		if i != len(scripts)-1 {
			fmt.Print(",")
		} else {
			fmt.Print("\n")
		}
	}

	return 0, nil
}

func (e *Executor) LoadAll() error {
	_, err := e.Load(context.Background(), []Script{
		luaZset.GetRevRange(),
		luaZset.GetCreate(),
		luaZset.GetGetField(),
		luaHash.GetCreate(),
		luaHash.GetGetField(),
		luaString.GetIncrBy(),
	})
	return err
}

func (e *Executor) Execute(ctx context.Context, script Script, keys []string, args ...interface{}) *redis.Cmd {
	return e.client.EvalSha(ctx, e.sha[script.Name()], keys, args)
}
