package util

type ZSetReadProcess interface {
	FindFromCache() (interface{}, bool)
	FindFromRedis() (interface{}, int)
	FindFromTiDB() (interface{}, error)
	ReBuildRedis()
}
type ZSetRead struct {
	ch chan ZSetReadProcess
}

func NewZSetRead(size int) *ZSetRead {
	zs := &ZSetRead{
		ch: make(chan ZSetReadProcess, size),
	}
	go func() {
		for p := range zs.ch {
			p.ReBuildRedis()
		}
	}()
	return zs
}

func (zs *ZSetRead) Read(p ZSetReadProcess) (interface{}, error) {
	resp, ok := p.FindFromCache()
	if ok {
		return resp, nil
	}
	resp, status := p.FindFromRedis()
	if status&ZSetFind == 0 {
		zs.ch <- p
		return p.FindFromTiDB()
	}
	if status&ZSetNeedRebuild != 0 {
		zs.ch <- p
	}
	if status&ZSetFindX != 0 || status&ZSetIsAll != 0 {
		return resp, nil
	}
	return p.FindFromTiDB()
}

const (
	ZSetFind        = 1 << 0 // ZSetMiss
	ZSetNeedRebuild = 1 << 1 // ZSetNeedNotRebuild
	ZSetIsAll       = 1 << 3 // ZSetNotAll
	ZSetFindX       = 1 << 4 // ZSetNotFindX

)
