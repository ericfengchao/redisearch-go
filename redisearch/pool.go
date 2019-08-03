package redisearch

import (
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

type ConnPool interface {
	Get() redis.Conn
	io.Closer
}

type singleHostPool struct {
	*redis.Pool
}

func NewSingleHostPool(host string) ConnPool {
	ret := redis.NewPool(func() (redis.Conn, error) {
		// TODO: Add timeouts. and 2 separate pools for indexing and querying, with different timeouts
		return redis.Dial("tcp", host)
	}, maxConns)
	ret.TestOnBorrow = func(c redis.Conn, t time.Time) (err error) {
		if time.Since(t) > time.Second {
			_, err = c.Do("PING")
		}
		return err
	}
	return &singleHostPool{ret}
}

type multiHostPool struct {
	sync.Mutex
	pools map[string]*redis.Pool
	hosts []string
}

func (p *multiHostPool) Close() error {
	p.Lock()
	defer p.Unlock()
	for _, pool := range p.pools {
		pool.Close()
	}
	return nil
}

func NewMultiHostPool(hosts []string) ConnPool {
	return &multiHostPool{
		pools: make(map[string]*redis.Pool, len(hosts)),
		hosts: hosts,
	}
}

func (p *multiHostPool) Get() redis.Conn {
	p.Lock()
	defer p.Unlock()
	host := p.hosts[rand.Intn(len(p.hosts))]
	pool, found := p.pools[host]
	if !found {
		pool = redis.NewPool(func() (redis.Conn, error) {
			// TODO: Add timeouts. and 2 separate pools for indexing and querying, with different timeouts
			return redis.Dial("tcp", host)
		}, maxConns)
		pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
			if time.Since(t).Seconds() > 1 {
				_, err := c.Do("PING")
				return err
			}
			return nil
		}

		p.pools[host] = pool
	}
	return pool.Get()

}
