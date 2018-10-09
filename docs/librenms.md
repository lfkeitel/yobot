# LibreNMS

***Module Type***: msgbus

***Internal/External***: internal

***Supports Aliases***: yes

## Description

The librenms module accepts alert events from LibreNMS.

## Configuration example

```toml
# Main module configuration, same as default route config
[routes.librenms]
Enabled  = true

[routes.librenms.settings]
address     = ""    # Address/hostname of LibreNMS server
skip_verify = false # Don't validate TLS certificates if any
apitoken    = ""    # LibreNMS API token

# Route sysContact information to a specific channel.
# This table is a map of email/contact information to a Mattermost channel.
# A key of "*" means a copy of all messages will be sent to that channel.
# An alert will go to all matching channels so order does not matter.
[routes.librenms.settings.routes]
"*"                 = "Global:NOC"        # Asterisk matches everything so all alerts will go here
"email@example.com" = "Server-Admins:NOC" # Alerts will go to this channel only if the email is in sysContact
```

## LibreNMS Configuration

In LibreNMS you will need to setup an API transport for alerts. The endpoint
follow the form `http://YOBOT_URL.example.com/msgbus/librenms?PARAMS`.
The following PARAMS need to be sent: title, host, sysName, severity, rule, and timestamp.

Example: `http://yobot.example.com/msgbus/librenms?title=%title&host=%hostname&sysName=%sysName&rule=%rulename&timestamp=%timestamp&severity=%severity`

Any extra parameters will be ignored. Missing parameters will be replaced with
a placeholder.

## Alert Routing

If `routes.librenms.settings.routes` is not defined, alerts are sent to the channels
per normal message bus rules.

If it is configured, the LibreNMS module handles the message routing instead.

The module will call the LibreNMS API and attempt to get information about the
alerted device. If the API call fails or the device isn't found, the module
falls back to normal message bus posting rules.

If a device is found, the sysContact information is checked against all configured
`settings.routes` keys. If a key is matched against the sysContact, a copy of the message
is sent to that channel. If the key is `*`, a copy of the message is sent regardless
of what the sysContact information is. The key can be a literal string or a regular
expression. Regular expressions can save space by consolidating multiple emails or
other strings into a single expression. The key doesn't need to be an email address
like the example. It can be any string. Emails are more typical however because
usually an email alert may also be sent.

Examples of keys:

- `*` - Matches any sysContact, an alert will always be sent to this channel (equivalent to `/.*/`).
- `some string` - Will match if "some string" is anywhere in sysContact.
- `/regexp?/` - Will match if the sysContact matches the regular expression.

It's recommended to configure ChannelOverride for the LibreNMS module just to
be sure that alerts are only going where you expect them to. Typically the
Channels settings and the default `[*.settings.routes]` channel will be the same.

```toml
[routes.librenms]
Enabled  = true
Channels = ["Global:NOC"] # This matches the catch-all channel below
ChannelOverride = true

[routes.librenms.settings.routes]
"*" = "Global:NOC"
"email@example.com" = "Server-Admins:NOC"
"/Servers.*?@example.com/" = "Server-Admins:NOC"
```
