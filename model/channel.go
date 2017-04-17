package model

import (
	"fmt"
	"strconv"
	"time"
)

// Channel represent use chat channel
type Channel struct {
	ChannelID int64
	Name      string
	Creator   int64
	Members   []Member
	CreatedAt time.Time
}

type Member struct {
	UserID int64
	JoinAt time.Time
}

func (c Channel) CreateChannel() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	c.ChannelID, err = redis.Int64(conn.Do("INCR", "next_channel_id"))
	if err != nil {
		return err
	}

	channelIDStr := strconv.FormatInt(c.ChannelID, 10)
	c.CreatedAt = time.Now().Unix()
	_, err = conn.Do("HMSET", "channel:"+channelIDStr, "name", c.Name, "creator", c.Creator, "created_at", c.CreatedAt)
	if err != nil {
		return err
	}

	userIDStr := strconv.FormatInt(c.Creator, 10)
	_, err = conn.Do("LPUSH", "channels:"+userIDStr, c.ChannelID)
	if err != nil {
		return err
	}

	userIDStr := strconv.FormatInt(c.Creator, 10)
	_, err = conn.Do("LPUSH", "channels:"+userIDStr, c.ChannelID)
	if err != nil {
		return err
	}
}

//JoinChannel
