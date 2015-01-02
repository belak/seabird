package plugins

import (
	"net/http"
	"bytes"
	"encoding/json"
	"github.com/belak/irc"
	"github.com/belak/seabird/bot"
	"github.com/belak/seabird/mux"
)

type ShortenResult struct {
	Kind string `json:"kind"`
	Id string `json:"id"`
	LongUrl string `json:"longUrl"`
}

func init() {
	bot.RegisterPlugin("tiny", NewTinyPlugin)
}

func NewTinyPlugin(m *mux.CommandMux) error {
	m.Event("tiny", Shorten)

	return nil
}

func Shorten(c *irc.Client, e *irc.Event) {
	if e.Trailing() == "" {
		c.MentionReply(e, "URL required")
		return
	}

	url := "https://www.googleapis.com/urlshortener/v1/url"

	var jsonStr = []byte(`{"longUrl":"` + e.Trailing() + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.MentionReply(e, "Error connecting to goo.gl")
	}
	defer resp.Body.Close()

	sr := new(ShortenResult)
	err = json.NewDecoder(resp.Body).Decode(sr)
	if err != nil {
		c.MentionReply(e, "Error reading server response")
	}

	c.MentionReply(e, sr.Id)
}