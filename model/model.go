package model

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRedisConn() (redis.Conn, error) {
	connectTime := time.Duration(1000 * 1000 * 1000)
	connectTimeoutOption := redis.DialConnectTimeout(connectTime)

	readTime := time.Duration(1000 * 1000 * 1000)
	readTimeoutOption := redis.DialReadTimeout(readTime)

	writeTime := time.Duration(1000 * 1000 * 1000)
	writeTimeoutOption := redis.DialWriteTimeout(writeTime)

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, errors.New("$REDIS_URL must be set")
	}

	return redis.DialURL(redisURL, connectTimeoutOption, readTimeoutOption, writeTimeoutOption)
}

func md5Password(password string) (salt, md5Password string) {
	salt = RandStringRunes(5)
	m5 := md5.New()
	m5.Write([]byte(password))
	m5.Write([]byte(string(salt)))
	st := m5.Sum(nil)
	md5Password = hex.EncodeToString(st)
	return
}

func Md5PasswordWithSalt(salt, password string) (md5Password string) {
	m5 := md5.New()
	m5.Write([]byte(password))
	m5.Write([]byte(string(salt)))
	st := m5.Sum(nil)
	md5Password = hex.EncodeToString(st)
	return
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandStringRunes create rand string
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func FlushDB() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("FLUSHDB")
	return err
}
