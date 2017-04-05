package model

import (
	"errors"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// User represent user　model
type User struct {
	UserID    int64    `json:"userid"`
	Username  string   `json:"username"`
	Salt      string   `json:"salt"`
	Password  string   `json:"password"`
	Auth      string   `json:"auth"`
	Followers []string `json:"followers"`
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
	u.Auth = authSecret

	_, err = conn.Do("HSET", "users", u.Username, userID)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", "auths", authSecret, userID)

	_, err = conn.Do("ZADD", "users_by_time", time.Now().Unix(), u.Username)
	return err
}

// GetUserByName used to get user info
func (u *User) GetUserByName() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	u.UserID, err = redis.Int64(conn.Do("HGET", "users", u.Username))
	if err != nil {
		return err
	}
	userIDStr := strconv.FormatInt(u.UserID, 10)

	u.Salt, err = redis.String(conn.Do("HGET", "user:"+userIDStr, "salt"))
	if err != nil {
		return err
	}

	u.Password, err = redis.String(conn.Do("HGET", "user:"+userIDStr, "password"))
	if err != nil {
		return err
	}

	u.Auth, err = redis.String(conn.Do("HGET", "user:"+userIDStr, "auth"))
	if err != nil {
		return err
	}

	u.Followers, err = redis.Strings(conn.Do("SMEMBERS", "followers:"+userIDStr))
	if err != nil {
		return err
	}

	u.Following, err = redis.Strings(conn.Do("SMEMBERS", "following:"+userIDStr))
	return err
}

// GetUserByAuth used to get userinfo
func (u *User) GetUserByAuth() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	u.UserID, err = redis.Int64(conn.Do("HGET", "auths", u.Auth))
	if err != nil {
		return err
	}
	userIDStr := strconv.FormatInt(u.UserID, 10)

	auth, err := redis.String(conn.Do("HGET", "user:"+userIDStr, "auth"))
	if err != nil {
		return err
	}

	if auth != u.Auth {
		return errors.New("auth fail")
	}

	u.Username, err = redis.String(conn.Do("HGET", "user:"+userIDStr, "username"))
	if err != nil {
		return err
	}

	u.Followers, err = redis.Strings(conn.Do("SMEMBERS", "followers:"+userIDStr))
	if err != nil {
		return err
	}

	u.Following, err = redis.Strings(conn.Do("SMEMBERS", "following:"+userIDStr))
	return err
}

//// UpdateUserAuth update user auth secret
//func (u *User) UpdateUserAuth() error {
//	authValue, err := R.Do("HGET", "user:"+u.UserID, "auth")
//	if err != nil {
//		return err
//	}

//	auth := authValue.(string)

//	newAuth := RandStringRunes(30)
//	_, err = R.Do("HSET", "user:"+u.UserID, "auth", newAuth)
//	if err != nil {
//		return err
//	}

//	_, err = R.Do("HSET", "auths", newAuth, u.UserID)
//	if err != nil {
//		return err
//	}

//	_, err = R.Do("HDEL", "auths", auth)
//	return err
//}

//// AddFollowers used to add followers
//func (u *User) AddFollowers(follwer User) (err error) {
//	followerAt := time.Now().Unix()
//	_, err = R.Do("ZADD", "followers:"+u.UserID, followerAt, follwer.UserID)
//	_, err = R.Do("ZADD", "following:"+follwer.UserID, followerAt, u.UserID)
//	return
//}

//// AddFollowingUser used to add following user
//func (u *User) AddFollowingUser(follwingUser User) (err error) {
//	followingAt := time.Now().Unix()
//	_, err = R.Do("ZADD", "following:"+u.UserID, followingAt, follwingUser.UserID)
//	_, err = R.Do("ZADD", "followers:"+follwingUser.UserID, followingAt, u.UserID)
//	return
//}
