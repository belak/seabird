# Seabird Configuration

Seabird has a single configuration file that combines general bot configuration options with plugin-specific options.

A sample config file is provided [here](../_extra/config.sample.toml). Note that
this config file only has values specified for plugins. Some may not be needed.

**How do I use a specific configuration file?**

Configuration is pulled from the environment variable `SEABIRD_CONFIG`. Set for a session with

```
export SEABIRD_CONFIG=$HOME/config.toml
```

To run one time with a specific configuration, run as:

```
SEABIRD_CONFIG=$HOME/config.toml go run cmd/seabird/main.go
```

**How do I enable a plugin?**

You can use the `plugins` configuration key to enable plugins:

```
$ cat _extra/config.sample.toml
...
plugins = [
  "db",
  "chance",
]
```

In this example the `db` and `chance` plugins are enabled.

**How can I enable a collection of similarly-named plugins?**

Similar to the previous point, you can use the `plugins` configuration key to enable a group of plugins using [glob syntax](https://github.com/gobwas/glob):

```
$ cat _extra/config.sample.toml
...
plugins = [
  "db",
  "url/*",
]
```

In this example the `db` is enabled, as well as all plugins whose names start with `"url/"`.

**What plugins are available in Seabird?**

Standard:

| Plugin | Description |
|--------|-------------||
|`bulkcnam` | This plugin is currently broken.|
|`chance` | This plugin adds support for flipping a coin and Russian Roulette.|
|`db` | This plugin adds support for other plugins to use a database. It doesn't expose commands.|
|`dice` | This plugin listens for [D&D](https://dnd.wizards.com)-style dice rolls and actions on them.|
|`fcc` | This plugin adds support for querying [HAM radio](http://www.arrl.org/what-is-ham-radio) licenses.|
|`forecast` | This plugin adds support for querying weather information for a location.|
|`google` | This plugin links to a Google search page for a specific query.|
|`issues` | This plugin adds support for Seabird GitHub issue search and creation.|
|`karma` | This plugin adds support for tracking karma points for things.|
|`lastseen` | This plugin adds support for showing when users last spoke in a channel.|
|`math` | This plugin adds basic support for mathematical expression evaluation.|
|`mentions` | This plugin enables Seabird to respond to certain fun messages.|
|`net_tools` | This plugin adds support for various network information commands.|
|`noaa` | This plugin adds support for querying [aviation weather information](https://en.wikipedia.org/wiki/METAR).|
|`phrases` | This plugin adds support for key-value tracking and delivery of user-defined phrases.|
|`remind` | This plugin adds support for user reminders.|
|`runescape` | This plugin adds support for querying [Old School RuneScape](https://oldschool.runescape.com) account information.|
|`stock` | This plugin adds support for querying stock information from the [IEX Cloud API](https://iexcloud.io/docs/api/).|
|`tiny` | This plugin adds support for shortening URLs.|
|`watchdog` | This plugin adds support for checking whether or not Seabird is alive and responding to commands.|
|`weight_tracker` | This plugin adds support for tracking weight over time.|
|`wikipedia` | This plugin adds support for querying Wikipedia.|

**What configuration options exist for Seabird?**

Configuration for the underlying [irc](gopkg.in/irc.v3) connection (see [irc.ClientConfig](https://godoc.org/gopkg.in/irc.v3#ClientConfig) for more information):

```
# User info
nick = "HelloWorld"
user = "seabird"
name = "seabird"
pass = "qwertyasdf"

pingfrequency = "10s"
pingtimeout = "10s"

sendlimit = "0"
sendburst = 4
```

Network connection information, used with either [net](https://golang.org/pkg/net/) or [tls](https://golang.org/pkg/crypto/tls/) depending on whether or not TLS is used:

```
# Combination of host and port to connect to
host = "chat.freenode.net:6697"

# Connect with TLS
tls = true

# From the Go docs:
#
# [tlsnoverify] controls whether a client verifies the
# server's certificate chain and host name.
# If [tlsnoverify] is true, TLS accepts any certificate
# presented by the server and any host name in that certificate.
# In this mode, TLS is susceptible to man-in-the-middle attacks.
# This should be used only for testing.
tlsnoverify false

# File paths for the X509 keypair to use when connecting with TLS
tlscert     "/path/to/certfile"
tlskey      "/path/to/keyfile"
```

IRC commands for the bot to send upon connecting:

```
cmds = [
  "JOIN #my-channel",
]
```

Command prefix for the bot, e.g. setting `prefix = "~"` would mean that you'd call a command named `foo` with a message like `~foo`.

```
prefix = "!"
```

As detailed above, `plugins` controls which plugins are enabled in the bot.

```
plugins = [
  "db",
  "runescape",
  "noaa",
]
```

`loglevel` controls the bot's log level. See [this](https://github.com/sirupsen/logrus/blob/master/logrus.go#L25) for supported levels. Note: `debug` has been deprecated. Don't use it.

```
loglevel = "info"
```

InfluxDB time series logging configuration:

```
# Enable or disable InfluxDB connections
enabled = true

# InfluxDB connection information
url = "my.influx.installation:1337"
username = "my_username"
password = "my_password"
database = "my_database"

# Precision of submitted datapoints. Realistically shouldn't be changed from "second".
precision = "s"

# Time interval to use when submitting points. Datapoints will be buffered up to
# `buffersize` before submitting.
submitinterval = "10s"

# Maximum number of points to queue before submitting to InfluxDB. Points
# gathered after this maximum size will be dropped.
buffersize = 50
```

For plugin-specific configuration see [this](./plugin_configuration_options.md).

[documentation index](./README.md)