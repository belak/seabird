package seabird

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	irc "gopkg.in/irc.v3"
)

//nolint:funlen
func TestCommandMux(t *testing.T) {
	// Empty mux should still have help
	mux := NewCommandMux("!")
	assert.Equal(t, 1, len(mux.cmdHelp))

	mh := &messageHandler{}

	b := &Bot{
		client: irc.NewClient(&bytes.Buffer{}, irc.ClientConfig{
			Nick: "bot",
			User: "herbert",
			Name: "Herbert Bot",
		}),
	}
	b.client.Run()

	// Ensure simple commands can be hit
	mux.Event("hello", mh.Handle, nil)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG #hello :!hello")))
	assert.Equal(t, 1, mh.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG bot :!hello")))
	assert.Equal(t, 2, mh.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG bot :hello")))
	assert.Equal(t, 3, mh.count)

	// Ensure command names are case insensitive
	mux = NewCommandMux("!")
	mh = &messageHandler{}

	mux.Event("hello", mh.Handle, nil)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG #hello :!hello")))
	assert.Equal(t, 1, mh.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG #hello :!Hello")))
	assert.Equal(t, 2, mh.count)

	// Ensure private commands don't work publicly
	mux = NewCommandMux("!")
	mh = &messageHandler{}

	mux.Private("hello", mh.Handle, nil)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG #hello :!hello")))
	assert.Equal(t, 0, mh.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG bot :!hello")))
	assert.Equal(t, 1, mh.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG bot :hello")))
	assert.Equal(t, 2, mh.count)

	// Ensure public commands don't work publicly
	mux = NewCommandMux("!")
	mh = &messageHandler{}

	mux.Channel("hello", mh.Handle, nil)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG #hello :!hello")))
	assert.Equal(t, 1, mh.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG bot :!hello")))
	assert.Equal(t, 1, mh.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG bot :hello")))
	assert.Equal(t, 1, mh.count)

	// Ensure commands are separate
	mux = NewCommandMux("!")
	mh = &messageHandler{}
	mh2 := &messageHandler{}

	mux.Event("hello1", mh.Handle, nil)
	mux.Event("hello2", mh2.Handle, nil)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG #hello :!hello1")))
	assert.Equal(t, 1, mh.count)
	assert.Equal(t, 0, mh2.count)
	mux.HandleEvent(b, NewRequest(irc.MustParseMessage(":belak PRIVMSG #hello :!hello2")))
	assert.Equal(t, 1, mh.count)
	assert.Equal(t, 1, mh2.count)
}
