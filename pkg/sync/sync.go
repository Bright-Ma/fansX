package syncx

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"time"
)

type Mutex struct {
	sync      *Sync
	key       string
	retry     int                           // 加锁最大重试次数
	value     string                        // 锁的value，标识谁加的锁
	delayFunc func(times int) time.Duration // 获取下次申请加锁的等待时间
	valueFunc func() string                 // 获取锁value的函数
	ttl       time.Duration                 // 锁过期时间
	keepalive float64                       // 保活系数，ttl*keepalive为保活间隔
	util      time.Duration                 // 最大保活时间
}

type Sync struct {
	client       *redis.Client
	unlockSha    string
	keepaliveSha string
}

// Lock 尝试加锁，直到达到最大重试次数
func (m *Mutex) Lock() error {
	return m.LockWithTimeout(0)
}

// LockWithTimeout 尝试加锁，直到超时或达到最大重试次数，0表示不设置超时时间
func (m *Mutex) LockWithTimeout(timeout time.Duration) error {
	var ctx context.Context
	var cancel context.CancelFunc
	// 无超时时间
	if timeout == 0 {
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	var ticker *time.Ticker

	for i := 0; i <= m.retry; i++ {
		if i == 0 {
			//成功
			if m.TryLock() == nil {
				return nil
			}
			// 防止引用空指针
			ticker = time.NewTicker(time.Hour)
			continue
		}
		// 设置等待时间
		ticker.Reset(m.delayFunc(i))

		select {
		// 超时
		case <-ctx.Done():
			ticker.Stop()
			return ErrTimeout
		case <-ticker.C:
			if m.TryLock() == nil {
				return nil
			}
		}
	}
	// 达到最大重试次数
	ticker.Stop()
	return ErrFailed
}

// TryLock 尝试加锁一次
func (m *Mutex) TryLock() error {
	value := m.valueFunc()
	ok, err := m.sync.client.SetNX(context.Background(), m.key, value, m.ttl).Result()
	// 其他错误
	if err != nil {
		return err
	}
	// 被其他实例锁住
	if !ok {
		return ErrFailed
	}
	m.value = value
	go func() {
		timeout, cancel := context.WithTimeout(context.Background(), m.util)
		defer cancel()
		client := m.sync.client
		sha := m.sync.keepaliveSha
		for {
			select {
			// 锁续期
			case <-time.After(time.Duration(float64(m.ttl) * m.keepalive)):
				res, err := client.EvalSha(context.Background(), sha, []string{m.key}, m.value, m.ttl).Result()
				if err != nil || res == nil {
					return
				}
			case <-timeout.Done():
				return
			}
		}
	}()
	return nil
}

// Unlock 解锁
func (m *Mutex) Unlock() error {
	res, err := m.sync.client.EvalSha(context.Background(), m.sync.unlockSha, []string{m.key}, m.value).Result()
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}
	if res.(string) == "also unlock" {
		return ErrAlsoUnlock
	}

	return ErrUnlockByOther
}

// NewMutex 并发不安全，加锁前科无限重试加锁，解锁后无法再次加锁
func (s *Sync) NewMutex(key string, options ...Option) *Mutex {
	mu := &Mutex{
		sync:  s,
		key:   key,
		retry: 50,
		value: "",
		// 重试时间间隔递增
		delayFunc: func(times int) time.Duration { return time.Duration((times/5+1)*(10+rand.Intn(20))) * time.Millisecond },
		valueFunc: func() string { return uuid.New().String() },
		ttl:       5 * time.Second,
		keepalive: 0.5,
		util:      time.Second * 60,
	}
	for _, option := range options {
		option.Apply(mu)
	}
	return mu
}

func NewSync(client *redis.Client) (*Sync, error) {
	sync := &Sync{client: client}
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := sync.LoadUnlock(timeout); err != nil {
		return nil, err
	}
	if err := sync.LoadKeepalive(timeout); err != nil {
		return nil, err
	}
	return sync, nil
}

func (s *Sync) LoadUnlock(ctx context.Context) error {
	sha, err := s.client.ScriptLoad(ctx, `
local key=KEYS[1]
local value=ARGV[1]

local res=redis.call("Get",key)

if res==nil then
    return "also unlock"
end

if res~=value then
    return "unlock by other"
end

redis.call("Del",key)

return nil
`).Result()
	if err != nil {
		return err
	}
	s.unlockSha = sha
	return nil
}

func (s *Sync) LoadKeepalive(ctx context.Context) error {
	sha, err := s.client.ScriptLoad(context.Background(), `
local key=KEYS[1]
local value=ARGV[1]
local ttl=ARGV[2]

local res=redis.call("Get",key)
if res==nil then
return nil
end

if res~=value then
return nil
end 

redis.call("Expire",key,ttl)

return 1
`).Result()
	if err != nil {
		return err
	}
	s.keepaliveSha = sha
	return nil
}

type Option interface {
	Apply(mutex *Mutex)
}

type OptionFunc func(mutex *Mutex)

func (f OptionFunc) Apply(mutex *Mutex) {
	f(mutex)
}

// WithRetry 设置最大重试次数，默认值50
func WithRetry(Retry int) OptionFunc {
	if Retry <= 0 {
		panic("invalid retry value")
	}
	return func(mutex *Mutex) {
		mutex.retry = Retry
	}
}

// WithDelayFunc 设置重试间隔函数，默认采取递增策略
func WithDelayFunc(f func(times int) time.Duration) OptionFunc {
	return func(mutex *Mutex) {
		mutex.delayFunc = f
	}
}

// WithValueFunc 设置value获取函数，默认为uuid
func WithValueFunc(f func() string) OptionFunc {
	return func(mutex *Mutex) {
		mutex.valueFunc = f
	}
}

// WithTTL 设置锁过期时间，默认为5s
func WithTTL(ttl time.Duration) OptionFunc {
	if ttl <= 0 {
		panic("invalid ttl value")
	}
	return func(mutex *Mutex) {
		mutex.ttl = ttl
	}
}

// WithKeepAlive 设置保活系数，默认0.5
func WithKeepAlive(keepalive float64) OptionFunc {
	if keepalive <= 0 || keepalive >= 1 {
		panic("invalid keepalive value")
	}
	return func(mutex *Mutex) {
		mutex.keepalive = keepalive
	}
}

// WithUtil 设置最大保活时间，默认60s
func WithUtil(util time.Duration) OptionFunc {
	if util <= 0 {
		panic("invalid util value")
	}
	return func(mutex *Mutex) {
		mutex.util = util
	}
}

var (
	ErrFailed        = errors.New("try lock failed")
	ErrTimeout       = errors.New("try lock timeout")
	ErrAlsoUnlock    = errors.New("also unlock")
	ErrUnlockByOther = errors.New("unlock by other")
)
