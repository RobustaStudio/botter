package main

import "github.com/go-redis/redis"

// a simple function to just connect to a redis-backend
func InitRedisClient(dsn string) (*redis.Client, error) {
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)
	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}
	go func() {
		client.FlushAll().Val()
	}()
	return client, nil
}
