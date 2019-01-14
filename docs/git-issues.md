# Git Issues

***Module Type***: msgbus

***Internal/External***: internal

***Supports Aliases***: yes

**Endpoint**: `/msgbus/git-issues`

## Description

The git module accepts webhook issue/comment events from Gitea.

## Configuration example

```toml
# Main module configuration, same as default route config
[routes.git-issues]
Enabled  = true

[routes.git-issues.settings]
secret = "" # Secret sent by GitHub/Gitea with each event
```

## Gitea Configuration

1. In the repo you want to send messages, go to Settings -> Webhooks.
2. Click Add Webhook, choose Gitea.
3. The payload URL will be `http://YOURSEVER/msgbus/git`.
4. Set Content Type to "application/json".
5. Generate a secret for the webhook. You will use this same secret in the
Yobot configuration file for the git module.
6. Under "When should this webhook be triggered?", choose "Custom Events...".
7. Check "Issues" and/or "Issue Comment"
8. Make sure "Active" is checked.
9. Click "Add Webhook".
