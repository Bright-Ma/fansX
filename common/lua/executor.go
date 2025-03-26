package lua

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Executor struct {
	client *redis.Client
	sha    map[string]string
}

func NewExecutor(client *redis.Client) *Executor {
	return &Executor{
		client: client,
	}
}

func (e *Executor) Load(ctx context.Context, scripts []Script) (int, error) {
	for i, script := range scripts {
		res, err := e.client.ScriptLoad(ctx, script.Function()).Result()
		if err != nil {
			return i + 1, err
		}
		e.sha[script.Name()] = res
	}

	return 0, nil
}

func (e *Executor) Execute(ctx context.Context, script Script, keys []string, args ...interface{}) *redis.Cmd {
	return e.client.EvalSha(ctx, e.sha[script.Name()], keys, args)
}
