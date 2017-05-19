package model

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

//'word_name':'' #单词
//'exchange': '' #单词的各种时态
//'symbols':'' #单词各种信息 下面字段都是这个字段下面的
//'ph_en': '' #英式音标
//'ph_am': '' #美式音标
//'ph_en_mp3':'' #英式发音
//'ph_am_mp3': '' #美式发音
//'ph_tts_mp3': '' #TTS发音
//'parts':'' #词的各种意思

type exchange struct {
	Pl    []string    `json:"word_pl"`
	Past  interface{} `json:"word_past,omitempty"`
	Donr  interface{} `json:"word_done,omitempty"`
	Ing   interface{} `json:"word_ing,omitempty"`
	Third interface{} `json:"word_third,omitempty"`
	Er    string      `json:"word_er"`
	Est   string      `json:"word_est"`
}

type part struct {
	Part  string   `json:"part"`
	Means []string `json:"means"`
}

type symbol struct {
	En     string `json:"ph_en"`
	Am     string `json:"ph_am"`
	Other  string `json:"ph_other"`
	Enmp3  string `json:"ph_en_mp3"`
	Ammp3  string `json:"ph_am_mp3"`
	Ttsmp3 string `json:"ph_tts_mp3"`
	Parts  []part `json:"parts"`
}

type Word struct {
	Name     string   `json:"word_name"`
	Exchange exchange `json:"exchange"`
	Symbols  []symbol `json:"symbols"`
	Items    []string `json:"items"`
}

type mean struct {
	Part string `json:"part"`
	Mean string `json:"mean"`
}

type ExpectWord struct {
	Name  string `json:"name"`
	Am    string `json:"am"`
	MP3   string `json:"mp3"`
	Means []mean `json:"means"`
}

func (w *Word) CreateWord() error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	url := "http://dict-co.iciba.com/api/dictionary.php?key=FB6A324454DE5CC292A2B1A893D2F2CA&type=json&w="
	url = url + w.Name
	r, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(w)

	if err != nil {
		return err
	}

	symbol := w.Symbols[0]
	parts := symbol.Parts

	_, err = conn.Do("HMSET", "word:"+w.Name, "am", symbol.Am, "ammp3", symbol.Ammp3)
	if err != nil {
		return err
	}

	for i := range parts {
		meanID, err := redis.Int64(conn.Do("INCR", "next_mean_id"))
		if err != nil {
			return err
		}

		meanIDStr := strconv.FormatInt(meanID, 10)

		part := parts[i]
		_, err = conn.Do("HMSET", "mean:"+meanIDStr, "part", part.Part, "mean", strings.Join(part.Means, ","))
		_, err = conn.Do("LPUSH", "means:"+w.Name, meanIDStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetWords(wordNames []string) ([]ExpectWord, error) {
	conn, err := getRedisConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ews := []ExpectWord{}
	for i := range wordNames {
		if !checkWordExist(wordNames[i], conn) {
			w := Word{Name: wordNames[i]}
			w.CreateWord()
		}

		ew := ExpectWord{}
		ew.Name = wordNames[i]
		// todo check err
		ew.Am, err = redis.String(conn.Do("HGET", "word:"+ew.Name, "am"))
		ew.MP3, err = redis.String(conn.Do("HGET", "word:"+ew.Name, "ammp3"))
		ms, err := redis.Strings(conn.Do("LRANGE", "means:"+ew.Name, 0, -1))
		if err != nil {
			return nil, err
		}
		means := []mean{}
		for i := range ms {
			m := mean{}
			meanIDStr := ms[i]
			m.Part, err = redis.String(conn.Do("HGET", "mean:"+meanIDStr, "part"))
			m.Mean, err = redis.String(conn.Do("HGET", "mean:"+meanIDStr, "mean"))
			means = append(means, m)
		}
		ew.Means = means
		ews = append(ews, ew)
	}
	return ews, nil
}

func checkWordExist(wordName string, conn redis.Conn) bool {
	et, err := redis.Int64(conn.Do("EXISTS", "word:"+wordName))
	if err != nil {
		return false
	}

	if et == 0 {
		return false
	}

	return true
}
