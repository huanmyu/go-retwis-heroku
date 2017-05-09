// word.go
package model

import (
	"encoding/json"
	"net/http"
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
	Pl    []string `json:"word_pl"`
	Past  []string `json:"word_past"`
	Donr  []string `json:"word_done"`
	Ing   []string `json:"word_ing"`
	Third []string `json:"word_third"`
	Er    string   `json:"word_er"`
	Est   string   `json:"word_est"`
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

func (w *Word) CreateWord() error {
	url := "http://dict-co.iciba.com/api/dictionary.php?w=go&key=FB6A324454DE5CC292A2B1A893D2F2CA&type=json"
	r, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(w)
	if err != nil {
		return err
	}
	return nil
}
