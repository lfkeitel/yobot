# Yobot Configuration File

Yobot's configuration file a single TOML file. There are sections to configure
the Mattermost server including login credentials and bot user details. Each message
bus route as a configuration for external connectivity and message routing.
Lastly, external modules can also have a configuration section. External module
configurations are specific to each module so you will need to consult the module's
documentation for available options. The included message bus modules all have
a common core configuration with a few having additional settings. For the additional
settings please consult the module docs.

## Main Section

```toml
[main]
Debug      = false
ExtraDebug = false
ModulesDir = "modules"
Modules    = []
DataDir    = "data"
```

`Debug` and `ExtraDebug` are used for development and should not be enabled in production.

`ModulesDir` is a directory that contains dynamically loaded modules.

`Modules` is an array of modules to load. The names of modules is their filename without
a `.so` extension. For example the module file `dandelion.so` would be loaded
with the name `dandelion`.

`DataDir` is a directory used by modules for their own data. Modules will store
data in a folder with the same name as the module.

## Mattermost

```toml
[mattermost]
Server       = ""
InsecureTLS  = false
DebugChannel = "Team:Channel"

[mattermost.login]
Username = ""
Password = ""

[mattermost.botid]
Firstname = ""
Lastname = ""
Nickname = ""
```

The mattermost section deals with connecting and interacting with the Mattermost
server. The main section defines the connection details. `Server` is the hostname
or IP address of Mattermost. `InsecureTLS` if true will make Yobot ignore invalid
TLS certificate details. The `DebugChannel` is used to interact with Yobot directly
for administrative tasks. Yobot will also post occasional messages to the channel
in case of errors or other problems. This channel should be private and only
accessible by the bot administrator and Yobot itself. There is no user checking
for admin commands.

The `mattermost.login` information is pretty self explanitory. Yobot expects to be a completely
separate user, not just a personal token on another user's account.

The `mattermost.botid` section specifies the display name and nickname of Yobot. Yobot will
update this information onces it connects to Mattermost if it doesn't match.

### Specifying Mattermost Channels

Throughout the Yobot configuration there are places where a Mattermost channel
needs to be configured. All channels are in the format `TeamName:ChannelName`.
Any spaces in either the team or channel name will be replaced with a dash
as Mattermost does when making either. Case does not matter. You can use the
team/channel display name or it's URL slug.

Once Yobot is running and it has posted at least one message to a particular
channel, changing a channel or team name will not require a restart. Yobot
caches the internal ID string after first use and uses that for later
posts. However, if you do change the name make sure to update the configuration
as soon as possible to match in case Yobot restarts.

## HTTP Server

```toml
[http]
Address = ":8080"
```

The `http` section configures Yobot's [message bus](message-bus.md). Yobot
does not currently support TLS encryption. If you would like that, Yobot
can run behind a forwarding proxy such as Nginx or Apache.

## Message Bus Routes

```toml
[routes.NAME]
Enabled         = true
Channels        = []
ChannelOverride = false
Username        = ""
Password        = ""
Alias           = ""

[routes.NAME.settings]
```

The message bus works by handling requests on specific endpoints. These endpoints
are handled by modules. Each message bus module has a `routes.` configuration
section.

- `Enabled` - Enable/disable the module.
- `Channels` - Extra channels in addition to the default route's channels to
send messages.
- `ChannelOverride` - Use the route's "Channels" setting exclusively instead of
merging with the default route's config.
- `Username` - HTTP basic auth username.
- `Password` - HTTP basic auth password.
- `Alias` - Make this route an alias for another.
- `Settings` - Custom configurations for a specific module. Consult the module's
docs to learn about these.

### Default Route

There's a special route called `default`. The default route doesn't actually
create the endpoint `/msgbus/default`. The settings for this route are used
for all routes unless overridden in the specific module's configuration section.
The exception to this rule is with channels. If a module's configuration has
the `Channels` setting, those channels are merged with the defaults. You can
change this by setting `ChannelOverride` to `true` in the module. With
`ChannelOverride`, the module's setting will be used instead of the default's.

### Route Aliases

A route alias uses the same underlying module but responds to a different
URL endpoint. This allows using different settings for different copies
of the same external application. See the [message bus docs](message-bus.md)
for more information.

## Plugin Modules

```toml
[modules.NAME]
...
```

A plugin module can also have a configuration section. These settings are specific
to the module and its documentation should be consulted for what they are.

## Example File

```toml
[main]

[mattermost]
Server = "https://localhost:8065"
InsecureTLS = false
DebugChannel = "Networking:yobot-test"

[mattermost.login]
Username = "yobot"
Password = "yobot"

[mattermost.botid]
Firstname = "Yobot"
Lastname = ""
Nickname = "yobot"

[http]
Address = ":8080"

[routes.default]
Channels = ["Networking:yobot-test"]

[routes.grafana] # Alerts
Enabled = true

[routes.general]
Enabled = true

[routes.librenms] # Alerts
Enabled = true

# Special module-handled routing
[routes.librenms.settings.routes]
"*" = "Global:NOC"
"email@domain.com" = "Server-Admins:NOC"

[routes.git] # Github/Gitea
Enabled = true

[routes.git.settings]
secret = "mysecret"

[modules.dandelion]
URL = "https://dandelion.example.com"
ApiKey = "123456789"
Channels = ["Networking:noc"]
```
