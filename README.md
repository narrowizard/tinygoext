# tinygoext
tinygo extension package.
+ redis session support for tinygo.

## usage
```go
import (
    "github.com/narrowizard/tinygoext/sessionext"
    "github.com/kdada/tinygo/session"
)

session.Register("redis",sessionext.NewRediSessionContainer)
```

## config
config in tinygo web.cfg  
+ SessionType redis
+ SessionSource 