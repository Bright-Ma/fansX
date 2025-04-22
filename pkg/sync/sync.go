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
	retry     int
	value     string
	delayFunc func(times int) time.Duration
	valueFunc func() string
	ttl       time.Duration
	keepalive float64
	util      time.Duration
}

type Sync struct {
	client       *redis.Client
	unlockSha    string
	keepaliveSha string
}

func (m *Mutex) Lock() error {
	return m.LockWithTimeout(0)
}

func (m *Mutex) LockWithTimeout(timeout time.Duration) error {
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout == 0 {
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	var ticker *time.Ticker

	for i := 0; i <= m.retry; i++ {
		if i == 0 {
			if m.TryLock() == nil {
				return nil
			}
			ticker = time.NewTicker(time.Hour)
			continue
		}
		ticker.Reset(m.delayFunc(i))

		select {
		case <-ctx.Done():
			ticker.Stop()
			return ErrTimeout
		case <-ticker.C:
			if m.TryLock() == nil {
				return nil
			}
		}
	}

	ticker.Stop()
	return ErrFailed
}

func (m *Mutex) TryLock() error {
	value := m.valueFunc()
	ok, err := m.sync.client.SetNX(context.Background(), m.key, value, m.ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrFailed
	}
	m.value = value
	go func() {
		client := m.sync.client
		sha := m.sync.keepaliveSha
		for {
			select {
			case <-time.After(time.Duration(float64(m.ttl) * m.keepalive)):
				res, err := client.EvalSha(context.Background(), sha, []string{m.key}, m.value, m.ttl).Result()
				if err != nil || res == nil {
					return
				}
			}
		}
	}()
	return nil
}

func (m *Mutex) Unlock() error {
	res, err := m.sync.client.EvalSha(context.Background(), m.sync.unlockSha, []string{m.key}, m.value).Result()
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}
	return errors.Join(ErrAlsoUnlock, errors.New(res.(string)))
}

func (s *Sync) NewMutex(key string, options ...Option) *Mutex {
	mu := &Mutex{
		sync:      s,
		key:       key,
		retry:     50,
		value:     "",
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

func WithRetry(Retry int) OptionFunc {
	if Retry <= 0 {
		panic("invalid retry value")
	}
	return func(mutex *Mutex) {
		mutex.retry = Retry
	}
}

func WithDelayFunc(f func(times int) time.Duration) OptionFunc {
	return func(mutex *Mutex) {
		mutex.delayFunc = f
	}
}

func WithValueFunc(f func() string) OptionFunc {
	return func(mutex *Mutex) {
		mutex.valueFunc = f
	}
}

func WithTTL(ttl time.Duration) OptionFunc {
	if ttl <= 0 {
		panic("invalid ttl value")
	}
	return func(mutex *Mutex) {
		mutex.ttl = ttl
	}
}

func WithKeepAlive(keepalive float64) OptionFunc {
	if keepalive <= 0 || keepalive >= 1 {
		panic("invalid keepalive value")
	}
	return func(mutex *Mutex) {
		mutex.keepalive = keepalive
	}
}

func WithUtil(util time.Duration) OptionFunc {
	if util <= 0 {
		panic("invalid util value")
	}
	return func(mutex *Mutex) {
		mutex.util = util
	}
}

var (
	ErrFailed     = errors.New("try lock failed")
	ErrTimeout    = errors.New("try lock timeout")
	ErrAlsoUnlock = errors.New("also unlock")
)
