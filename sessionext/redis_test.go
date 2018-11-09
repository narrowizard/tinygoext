package sessionext

import (
	"testing"
)

var source = `{"Host":"127.0.0.1:6379","MaxIdle":10,"MaxActive":10,"IdleTimeout":60,"Wait":false,"DB":2,"Password":""}`

func TestRediSessionContainer(t *testing.T) {
	var container, err = NewRediSessionContainer(1800, source)
	if err != nil {
		t.Error(err)
	}
	var session, succ = container.CreateSession()
	if !succ {
		t.Error("create session failed!")
	}
	session.SetString("test", "value1")
	session, succ = container.Session(session.SessionId())
	if !succ {
		t.Error("get session failed!")
	}
	val, succ := session.String("test")
	if !succ {
		t.Error("get string failed!")
	}
	if val != "value1" {
		t.Error("string value error!")
	}
}
