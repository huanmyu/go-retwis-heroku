package model

import (
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Post represent post model
type Post struct {
	PostID    int64  `json:"postid"`
	UserID    int64  `json:"userid"`
	CreatedAt int64  `json:"created_at"`
	Content   string `json:"content"`
}

// CreatePost used to create post
func (p *Post) CreatePost() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	p.PostID, err = redis.Int64(conn.Do("INCR", "next_post_id"))
	if err != nil {
		return err
	}

	postIDStr := strconv.FormatInt(p.PostID, 10)
	p.CreatedAt = time.Now().Unix()

	_, err = conn.Do("HMSET", "post:"+postIDStr, "userid", p.UserID, "created_at", p.CreatedAt, "content", p.Content)
	if err != nil {
		return err
	}

	userIDStr := strconv.FormatInt(p.UserID, 10)
	followerIDs, err := redis.Strings(conn.Do("ZRANGE", "followers:"+userIDStr, 0, -1))
	if err != nil {
		return err
	}

	for i := range followerIDs {
		followerID := followerIDs[i]
		_, err = conn.Do("LPUSH", "posts:"+followerID, p.PostID)
		if err != nil {
			return err
		}
	}

	_, err = conn.Do("LPUSH", "posts:"+userIDStr, p.PostID)
	if err != nil {
		return err
	}

	_, err = conn.Do("LPUSH", "timeline", p.PostID)
	if err != nil {
		return err
	}

	_, err = conn.Do("LTRIM", "timeline", 0, 1000)
	return err
}

// GetPost used to get post
func (p *Post) GetPost(postID string) error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	p.CreatedAt, err = redis.Int64(conn.Do("HGET", "post:"+postID, "created_at"))
	if err != nil {
		return err
	}

	p.Content, err = redis.String(conn.Do("HGET", "post:"+postID, "content"))
	return err
}

// GetTimelinePosts used to get timeline posts
func (p *Post) GetTimelinePosts() ([]string, error) {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	return redis.Strings(conn.Do("LRANGE", "timeline", 0, 50))
}
