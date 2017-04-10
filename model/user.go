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
	PostIDs   []string `json:"postids"`
}

// CreateUser used to create user
func (u *User) CreateUser() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	isExist, err := redis.Bool(conn.Do("SISMEMBER", "usernames", u.Username))
	if err != nil {
		return err
	}

	if isExist {
		return errors.New("username already exists")
	}

	userID, err := redis.Int64(conn.Do("INCR", "next_user_id"))
	if err != nil {
		return err
	}

	salt, password := md5Password(u.Password)
	authSecret, err := CreateUUID()
	if err != nil {
		authSecret = RandStringRunes(30)
	}

	_, err = conn.Do("HMSET", "user:"+strconv.FormatInt(userID, 10), "username", u.Username, "salt", salt, "password", password, "auth", authSecret)
	if err != nil {
		return err
	}
	u.Auth = authSecret

	_, err = conn.Do("SADD", "usernames", u.Username)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", "users", u.Username, userID)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", "auths", authSecret, userID)

	_, err = conn.Do("ZADD", "users_by_time", time.Now().Unix(), u.Username)
	return err
}

func (u *User) GetUserByUserID() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	userIDStr := strconv.FormatInt(u.UserID, 10)
	u.Username, err = redis.String(conn.Do("HGET", "user:"+userIDStr, "username"))
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

	u.Followers, err = redis.Strings(conn.Do("ZRANGE", "followers:"+userIDStr, 0, -1))
	if err != nil {
		return err
	}

	u.Following, err = redis.Strings(conn.Do("ZRANGE", "following:"+userIDStr, 0, -1))
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

	u.Followers, err = redis.Strings(conn.Do("ZRANGE", "followers:"+userIDStr, 0, -1))
	if err != nil {
		return err
	}

	u.Following, err = redis.Strings(conn.Do("ZRANGE", "following:"+userIDStr, 0, -1))
	return err
}

// UpdateUserAuth update user auth secret
func (u *User) UpdateUserAuth() error {
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

	newAuth, err := CreateUUID()
	if err != nil {
		newAuth = RandStringRunes(30)
	}

	_, err = conn.Do("HSET", "user:"+userIDStr, "auth", newAuth)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", "auths", newAuth, u.UserID)
	if err != nil {
		return err
	}

	_, err = conn.Do("HDEL", "auths", u.Auth)
	return err
}

// GetUserPosts used to get user postids
func (u *User) GetUserPosts(start int, count int) error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	userIDStr := strconv.FormatInt(u.UserID, 10)
	u.PostIDs, err = redis.Strings(conn.Do("LRANGE", "posts:"+userIDStr, start, start+count))
	return err
}

func (u *User) GetUserPostCount() (int64, error) {
	conn, err := getRedisConn()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	userIDStr := strconv.FormatInt(u.UserID, 10)
	return redis.Int64(conn.Do("LLEN", "posts:"+userIDStr))

}

func (u *User) GetLastUsers() ([]string, error) {
	conn, err := getRedisConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return redis.Strings(conn.Do("ZREVRANGE", "users_by_time", 0, 9))
}

func (u *User) IsFollowing(follwing *User) (bool, error) {
	conn, err := getRedisConn()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	userIDStr := strconv.FormatInt(u.UserID, 10)
	score, err := redis.Int64(conn.Do("ZSCORE", "following:"+userIDStr, follwing.UserID))
	if err != nil {
		return false, err
	}

	if score > 0 {
		return true, nil
	}

	return false, nil
}

// AddFollowingUser used to add following user
func (u *User) AddOrRemFollowingUser(followingUser User, following string) error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	userIDStr := strconv.FormatInt(u.UserID, 10)
	followingUserIDStr := strconv.FormatInt(followingUser.UserID, 10)
	if following == "1" {
		followingAt := time.Now().Unix()
		_, err = conn.Do("ZADD", "following:"+userIDStr, followingAt, followingUser.UserID)
		_, err = conn.Do("ZADD", "followers:"+followingUserIDStr, followingAt, u.UserID)
	} else {
		_, err = conn.Do("ZREM", "following:"+userIDStr, followingUser.UserID)
		_, err = conn.Do("ZREM", "followers:"+followingUserIDStr, u.UserID)
	}

	return err
}
