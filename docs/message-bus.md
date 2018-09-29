# Message Bus

The message bus is the primary, builtin method for external applications
to send alerts/events to Yobot. There's a main `/msgbus/` endpoint that modules
can attach to and handle events based on the next path segment. For example
`/msgbus/git` is handled by the git module. Modules can also mount themselves
to a different base path but then lose a few automatic features such as aliasing
and HTTP basic authentication.

## Developer API
