package model

import (
	"strconv"

	"github.com/garyburd/redigo/redis"
)

// User represent user　model
type User struct {
	UserID    int64    `json:"userid"`
	Username  string   `json:"username"`
	Slat      string   `json:"slat"`
	Password  string   `json:"password"`
	Auth      string   `json:"auth"`
	Follows   []string `json:"follows"`
	Following []string `json:"following"`
}

// CreateUser used to create user
func (u *User) CreateUser() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	userID, err := redis.Int64(conn.Do("INCR", "next_user_id"))
	if err != nil {
		return err
	}

	salt, password := md5Password(u.Password)
	authSecret := RandStringRunes(30)

	_, err = conn.Do("HMSET", "user:"+strconv.FormatInt(userID, 10), "username", u.Username, "salt", salt, "password", password, "auth", authSecret)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", "users", u.Username, userID)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", "auths", authSecret, userID)
	return err
}
