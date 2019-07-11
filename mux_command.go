package seabird

import (
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"

	irc "gopkg.in/irc.v3"
)

// CommandMux is a simple IRC event multiplexer, based on the BasicMux.

// HelpInfo is a collection of instructions for command usage that
// is formatted with <prefix>help
type HelpInfo struct {
	name        string
	Usage       string
	Description string
}

// The CommandMux is given a prefix string and matches all PRIVMSG
// events which start with it. The first word after the string is
// moved into the Event.Command.
type CommandMux struct {
	private        *BasicMux
	public         *BasicMux
	prefix         string
	cmdHelp        map[string]*HelpInfo
	displaySimilar bool
}

// NewCommandMux will create an initialized BasicMux with no handlers.
func NewCommandMux(prefix string, displaySimilar bool) *CommandMux {
	m := &CommandMux{
		NewBasicMux(),
		NewBasicMux(),
		prefix,
		make(map[string]*HelpInfo),
		displaySimilar,
	}
	m.Event("help", m.help, &HelpInfo{
		"help",
		"<command>",
		"Displays help messages for a given command",
	})
	return m
}

func (m *CommandMux) help(b *Bot, msg *irc.Message) {
	cmd := msg.Trailing()
	if cmd == "" {
		// Get all keys
		keys := make([]string, 0, len(m.cmdHelp))
		for k := range m.cmdHelp {
			keys = append(keys, k)
		}

		// Sort everything
		sort.Strings(keys)

		if b.FromChannel(msg) {
			// If they said "!help" in a channel, list all available commands
			b.Reply(msg, "Available commands: %s. Use %shelp [command] for more info.", strings.Join(keys, ", "), m.prefix)
		} else {
			for _, v := range keys {
				h := m.cmdHelp[v]
				if h.Usage != "" {
					b.Reply(msg, "%s %s: %s", v, h.Usage, h.Description)
				} else {
					b.Reply(msg, "%s: %s", v, h.Description)
				}
			}
		}
	} else if help, ok := m.cmdHelp[cmd]; ok {
		if help == nil {
			b.Reply(msg, "There is no help available for command %q", cmd)
		} else {
			lines := help.format(m.prefix, cmd)
			for _, line := range lines {
				b.Reply(msg, "%s", line)
			}
		}
	} else {
		b.MentionReply(msg, "There is no help available for command %q", cmd)
	}
}

func (h *HelpInfo) format(prefix, command string) []string {
	if h.Usage == "" && h.Description == "" {
		return []string{"There is no help available for command " + command}
	}

	ret := []string{}

	if h.Usage != "" {
		ret = append(ret, "Usage: "+prefix+h.name+" "+h.Usage)
	}

	if h.Description != "" {
		ret = append(ret, h.Description)
	}

	return ret
}

// Event will register a Handler as both a private and public command
func (m *CommandMux) Event(c string, h HandlerFunc, help *HelpInfo) {
	if help != nil {
		help.name = c
	}
	c = strings.ToLower(c)

	m.private.Event(c, h)
	m.public.Event(c, h)

	m.cmdHelp[c] = help
}

// Channel will register a handler as a public command
func (m *CommandMux) Channel(c string, h HandlerFunc, help *HelpInfo) {
	if help != nil {
		help.name = c
	}
	c = strings.ToLower(c)

	m.public.Event(c, h)

	m.cmdHelp[c] = help
}

// Private will register a handler as a private command
func (m *CommandMux) Private(c string, h HandlerFunc, help *HelpInfo) {
	if help != nil {
		help.name = c
	}
	c = strings.ToLower(c)

	m.private.Event(c, h)

	m.cmdHelp[c] = help
}

// HandleEvent strips off the prefix, pulls the command out
// and runs HandleEvent on the internal BasicMux
func (m *CommandMux) HandleEvent(b *Bot, msg *irc.Message) {
	if msg.Command != "PRIVMSG" {
		// TODO: Log this
		return
	}

	// Get the last arg and see if it starts with the command prefix
	lastArg := msg.Trailing()
	if b.FromChannel(msg) && !strings.HasPrefix(lastArg, m.prefix) {
		return
	}

	// Copy it into a new Event
	newEvent := msg.Copy()

	// Chop off the command itself
	msgParts := strings.SplitN(lastArg, " ", 2)
	newEvent.Params[len(newEvent.Params)-1] = ""
	if len(msgParts) > 1 {
		newEvent.Params[len(newEvent.Params)-1] = strings.TrimSpace(msgParts[1])
	}

	newEvent.Command = strings.ToLower(msgParts[0])
	if strings.HasPrefix(newEvent.Command, m.prefix) {
		newEvent.Command = newEvent.Command[len(m.prefix):]
	}

	var commands []string
	var distances []int
	if b.FromChannel(newEvent) {
		m.public.HandleEvent(b, newEvent)
		commands = m.public.eventNames()
	} else {
		m.private.HandleEvent(b, newEvent)
		commands = m.public.eventNames()
	}

	if m.displaySimilar {
		for _, v := range commands {
			distances = append(distances, levenshtein.ComputeDistance(newEvent.Command, v))
		}

		minDistance := minIntSlice(distances)

		// 0 means it matched exactly so we don't need recommendations
		if minDistance == 0 {
			return
		}

		var similarCommands []string
		for k, v := range distances {
			if v-minDistance < 1 {
				similarCommands = append(similarCommands, commands[k])
			}
		}

		sort.Strings(similarCommands)

		if len(similarCommands) > 0 {
			b.MentionReply(msg, "Did you mean one of the following: %s", strings.Join(similarCommands, ", "))
		}
	}
}
