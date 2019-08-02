package url

import (
	"net/url"
	"regexp"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/lrstanley/girc"

	seabird "github.com/belak/go-seabird"
)

func init() {
	seabird.RegisterPlugin("url/twitter", newtwitterProvider)
}

type twitterConfig struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

type twitterProvider struct {
	api *anaconda.TwitterApi
}

var twitterStatusRegex = regexp.MustCompile(`^/.*?/status/(.+)$`)
var twitterUserRegex = regexp.MustCompile(`^/([^/]+)$`)
var twitterPrefix = "[Twitter]"

func newtwitterProvider(b *seabird.Bot, urlPlugin *Plugin) error {
	t := &twitterProvider{}

	tc := &twitterConfig{}
	err := b.Config("twitter", tc)
	if err != nil {
		return err
	}

	anaconda.SetConsumerKey(tc.ConsumerKey)
	anaconda.SetConsumerSecret(tc.ConsumerSecret)
	t.api = anaconda.NewTwitterApi(tc.AccessToken, tc.AccessTokenSecret)

	urlPlugin.RegisterProvider("twitter.com", t.Handle)

	return nil
}

func (t *twitterProvider) Handle(c *girc.Client, e girc.Event, u *url.URL) bool {
	if twitterUserRegex.MatchString(u.Path) {
		return t.getUser(c, e, u.Path)
	} else if twitterStatusRegex.MatchString(u.Path) {
		return t.getTweet(c, e, u.Path)
	}

	return false
}

func (t *twitterProvider) getUser(c *girc.Client, e girc.Event, url string) bool {
	matches := twitterUserRegex.FindStringSubmatch(url)
	if len(matches) != 2 {
		return false
	}

	user, err := t.api.GetUsersShow(matches[1], nil)
	if err != nil {
		return false
	}

	// Jay Vana (@jsvana) - Description description
	c.Cmd.Replyf(e, "%s %s (@%s) - %s", twitterPrefix, user.Name, user.ScreenName, user.Description)

	return true
}

func (t *twitterProvider) getTweet(c *girc.Client, e girc.Event, url string) bool {
	matches := twitterStatusRegex.FindStringSubmatch(url)
	if len(matches) != 2 {
		return false
	}

	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return false
	}

	tweet, err := t.api.GetTweet(id, nil)
	if err != nil {
		return false
	}

	// Tweet text (@jsvana)
	c.Cmd.Replyf(e, "%s %s (@%s)", twitterPrefix, tweet.Text, tweet.User.ScreenName)

	return true
}
