# tinygoext
tinygo extension package.
+ redis session support for tinygo.

## usage
```go
import (
    "github.com/kdada/tinygo/session"
	"github.com/narrowizard/tinygoext/sessionext"
)

session.Register("redis", sessionext.NewRediSessionContainer)
```

## config
config in tinygo web.cfg  
+ SessionType redis
+ SessionSource {"Host":"localhost:6379","MaxIdle":10,"MaxActive":20,"IdleTimeout":60,"Wait":false,"DB":2,"Password":""}