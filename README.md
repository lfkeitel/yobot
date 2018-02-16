# Yobot

Yobot is a simple appliction that takes messages from HTTP and sends them to
IRC channels. Yobot allows anything that can send simple HTTP and JSON to send
alerts or other messages to and IRC server.

## Default Handlers

All handlers requires HTTP POST method.

### /msgbus/grafana

For grafana webhooks

### /msgbus/general

For general application. Expected content:

```json
{
    "title": "This is a title",
    "message": "This is a message"
}
```

### /msgbus/librenms

Uses URL parameters:

- `host` - Hostname/system name
- `title` - Title of alert
- `rule` - Rule name
- `severity` - Alert severity
- `timestamp` - Alert timestamp

LibreNMS doesn't support basic authentication for the API transport. Enabling
basic authentication on this route (or aliased routes) will not work.

### /msgbus/git

For gitea webhooks. Gitea doesn't support HTTP basic auth but does allow setting
a secret in the message. The secret and be configured using the route settings.

```toml
[routes.git.settings]
secret = "mysecret"
```
