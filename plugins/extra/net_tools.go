package extra

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/Unknwon/com"
	ping "github.com/belak/go-ping"
	"github.com/belak/go-seabird"
	irc "github.com/go-irc/irc"
)

func init() {
	seabird.RegisterPlugin("nettools", newNetToolsPlugin)
}

type netToolsPlugin struct {
	Key            string
	PrivilegedPing bool
}

func newNetToolsPlugin(b *seabird.Bot, cm *seabird.CommandMux) error {
	p := &netToolsPlugin{}

	err := b.Config("net_tools", p)
	if err != nil {
		return err
	}

	cm.Event("rdns", p.RDNS, &seabird.HelpInfo{
		Usage:       "<ip>",
		Description: "Does a reverse DNS lookup on the given IP",
	})
	cm.Event("dig", p.Dig, &seabird.HelpInfo{
		Usage:       "<domain>",
		Description: "Retrieves IP records for given domain",
	})
	cm.Event("ping", p.Ping, &seabird.HelpInfo{
		Usage:       "<host>",
		Description: "Pings given host once",
	})
	cm.Event("traceroute", p.Traceroute, &seabird.HelpInfo{
		Usage:       "<host>",
		Description: "Runs traceroute on given host and returns pastebin URL for results",
	})
	cm.Event("whois", p.Whois, &seabird.HelpInfo{
		Usage:       "<domain>",
		Description: "Runs whois on given domain and returns pastebin URL for results",
	})
	cm.Event("dnscheck", p.DNSCheck, &seabird.HelpInfo{
		Usage:       "<domain>",
		Description: "Returns DNSCheck URL for domain",
	})
	cm.Event("asn", p.ASNLookup, &seabird.HelpInfo{
		Usage:       "<ip>",
		Description: "Return subnet info for a given IP",
	})

	return nil
}

func (p *netToolsPlugin) RDNS(b *seabird.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Argument required")
			return
		}
		names, err := net.LookupAddr(m.Trailing())
		if err != nil {
			b.MentionReply(m, err.Error())
			return
		}

		if len(names) == 0 {
			b.MentionReply(m, "No results found")
			return
		}

		b.MentionReply(m, names[0])

		if len(names) > 1 {
			for _, name := range names[1:] {
				b.Writef("NOTICE %s :%s", m.Prefix.Name, name)
			}
		}
	}()
}

func (p *netToolsPlugin) Dig(b *seabird.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Domain required")
			return
		}

		addrs, err := net.LookupHost(m.Trailing())
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		if len(addrs) == 0 {
			b.MentionReply(m, "No results found")
			return
		}

		b.MentionReply(m, addrs[0])

		if len(addrs) > 1 {
			for _, addr := range addrs[1:] {
				b.Writef("NOTICE %s :%s", m.Prefix.Name, addr)
			}
		}
	}()
}

func (p *netToolsPlugin) Ping(b *seabird.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Host required")
			return
		}

		pinger, err := ping.NewPinger(m.Trailing())
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}
		pinger.Count = 1
		pinger.SetPrivileged(p.PrivilegedPing)

		pinger.OnRecv = func(pkt *ping.Packet) {
			b.MentionReply(m, "%d bytes from %s: icmp_seq=%d time=%s",
				pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
		}
		err = pinger.Run()
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}
	}()
}

func (p *netToolsPlugin) pasteData(data string) (string, error) {
	resp, err := http.PostForm("http://pastebin.com/api/api_post.php", url.Values{
		"api_dev_key":    {p.Key},
		"api_option":     {"paste"},
		"api_paste_code": {data},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

func (p *netToolsPlugin) runCommand(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "", err
	}

	return p.pasteData(string(out))
}

func (p *netToolsPlugin) handleCommand(b *seabird.Bot, m *irc.Message, command string, emptyMsg string) {
	if m.Trailing() == "" {
		b.MentionReply(m, "Host required")
		return
	}

	url, err := p.runCommand("traceroute", m.Trailing())
	if err != nil {
		b.MentionReply(m, "%s", err)
		return
	}

	b.MentionReply(m, "%s", url)

}

func (p *netToolsPlugin) Traceroute(b *seabird.Bot, m *irc.Message) {
	go p.handleCommand(b, m, "traceroute", "Host required")
}

func (p *netToolsPlugin) Whois(b *seabird.Bot, m *irc.Message) {
	go p.handleCommand(b, m, "whois", "Domain required")
}

func (p *netToolsPlugin) DNSCheck(b *seabird.Bot, m *irc.Message) {
	if m.Trailing() == "" {
		b.MentionReply(m, "Domain required")
		return
	}

	b.MentionReply(m, "https://www.whatsmydns.net/#A/"+m.Trailing())
}

type asnResponse struct {
	Announced     bool
	AsCountryCode string `json:"as_country_code"`
	AsDescription string `json:"as_description"`
	AsNumber      int    `json:"as_number"`
	FirstIP       string `json:"first_ip"`
	LastIP        string `json:"last_ip"`
}

func (p *netToolsPlugin) ASNLookup(b *seabird.Bot, m *irc.Message) {
	if m.Trailing() == "" {
		b.MentionReply(m, "IP required")
		return
	}

	asnResp := asnResponse{}

	err := com.HttpGetJSON(
		&http.Client{},
		"https://api.iptoasn.com/v1/as/ip/"+m.Trailing(),
		&asnResp)
	if err != nil {
		b.MentionReply(m, "%s", err)
		return
	}

	if !asnResp.Announced {
		b.MentionReply(m, "ASN information not available")
		return
	}

	b.MentionReply(
		m,
		"#%d (%s - %s) - %s (%s)",
		asnResp.AsNumber,
		asnResp.FirstIP,
		asnResp.LastIP,
		asnResp.AsDescription,
		asnResp.AsCountryCode)
}
