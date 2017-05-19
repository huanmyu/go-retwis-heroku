package main

import (
	"github.com/garyburd/redigo/redis"
	"testing"
	"time"
)

func TestRedisConn(t *testing.T) {
	conn, err := getRedisConn()
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	s, err := redis.String(conn.Do("ping"))
	if err != nil {
		t.Error(err)
	}
	if s != "PONG" {
		t.Errorf("want pong not %q", s)
	}
}

func getRedisConn() (redis.Conn, error) {
	connectTime := time.Duration(1000 * 1000 * 1000)
	connectTimeoutOption := redis.DialConnectTimeout(connectTime)

	readTime := time.Duration(1000 * 1000 * 1000)
	readTimeoutOption := redis.DialReadTimeout(readTime)

	writeTime := time.Duration(1000 * 1000 * 1000)
	writeTimeoutOption := redis.DialWriteTimeout(writeTime)

	redisURL := "redis://localhost:6379"

	return redis.DialURL(redisURL, connectTimeoutOption, readTimeoutOption, writeTimeoutOption)
}
