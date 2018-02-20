package rediscache

import (
	"context"
	"github.com/go-redis/redis"
	"golang.org/x/crypto/acme/autocert"
)

// Logger logs.
type Logger interface {
	Printf(format string, v ...interface{})
}

type Cache struct {
	RedisURL string
	Client   *redis.Client
	Logger   Logger
}

var _ autocert.Cache = (*Cache)(nil)

func New(redisURL string) (*Cache, error) {

	opts, err := redis.ParseURL(redisURL)
	client := redis.NewClient(opts)

	if err != nil {
		return nil, err
	}

	resp := client.Ping()
	err = resp.Err()
	if err != nil {
		return nil, err
	}

	return &Cache{
		RedisURL: redisURL,
		Client:   client,
	}, nil
}

func (c *Cache) log(format string, v ...interface{}) {
	if c.Logger == nil {
		return
	}
	c.Logger.Printf(format, v)
}

func (c *Cache) get(key string) ([]byte, error) {
	resp := c.Client.Get(key)
	data, err := resp.Bytes()
	return data, err
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	c.log("Cache Get %s", key)

	var (
		data []byte
		err  error
		done = make(chan struct{})
	)

	go func() {
		data, err = c.get(key)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}

	if err != nil {
		return nil, autocert.ErrCacheMiss
	}

	return data, err
}

func (c *Cache) delete(key string) error {
	resp := c.Client.Del(key)
	return resp.Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	c.log("Cache Delete %s", key)

	var (
		err  error
		done = make(chan struct{})
	)

	go func() {
		err = c.delete(key)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	return err
}

func (c *Cache) put(key, data string) error {
	resp := c.Client.Set(key, data, 0)
	return resp.Err()

}

func (c *Cache) Put(ctx context.Context, key string, data []byte) error {
	c.log("Cache Put %s", key)

	var (
		err  error
		done = make(chan struct{})
	)

	go func() {
		err = c.put(key, string(data))
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	return err
}
