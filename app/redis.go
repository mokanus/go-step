package app

import (
	"errors"
	"github.com/gomodule/redigo/redis"
)

type RedisConn struct {
	addr string
	pool *redis.Pool
}

func RedisDo(commandName string, args ...interface{}) (interface{}, error) {
	if defaultRedisConn == nil {
		return nil, errors.New("没有可用的Redis服务")
	}
	c := defaultRedisConn.pool.Get()
	defer c.Close()
	return c.Do(commandName, args...)
}

func RedisString(commandName string, args ...interface{}) (string, error) {
	if defaultRedisConn == nil {
		return "", errors.New("没有可用的Redis服务")
	}

	c := defaultRedisConn.pool.Get()
	defer c.Close()

	reply, err := c.Do(commandName, args...)
	if err != nil {
		return "", err
	}

	if reply == nil {
		return "", nil
	}

	return redis.String(reply, nil)
}

func RedisStringMap(commandName string, args ...interface{}) (map[string]string, error) {
	if defaultRedisConn == nil {
		return nil, errors.New("没有可用的Redis服务")
	}

	c := defaultRedisConn.pool.Get()
	defer c.Close()

	reply, err := c.Do(commandName, args...)
	if err != nil {
		return nil, err
	}

	if reply == nil {
		return nil, nil
	}

	return redis.StringMap(reply, nil)
}
