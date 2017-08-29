package session

import (
	"github.com/garyburd/redigo/redis"
	"github.com/kdada/tinygo/session"
	"github.com/kdada/tinygo/util"
)

type RediSession struct {
	sessionId string                 // 会话id
	data      map[string]interface{} // 数据
}

func newRediSession(sessionId string) *RediSession {
	var ss = new(RediSession)
	ss.sessionId = sessionId
	ss.data = make(map[string]interface{}, 0)
	return ss
}

func (this *RediSession) SessionId() string {
	return this.sessionId
}

func (this *RediSession) Value(key string) (interface{}, bool) {
	return nil, false
}

func (this *RediSession) String(key string) (string, bool) {
	return "", false
}

func (this *RediSession) Int(key string) (int, bool) {
	return 0, false
}

func (this *RediSession) Bool(key string) (bool, bool) {
	return false, false
}

func (this *RediSession) Float(key string) (float64, bool) {
	return 0, false
}

func (this *RediSession) SetValue(key string, value interface{}) {
	return
}

func (this *RediSession) SetString(key string, value string) {
	return
}

func (this *RediSession) SetInt(key string, value int) {
	return
}

func (this *RediSession) SetBool(key string, value bool) {
	return
}

func (this *RediSession) SetFloat(key string, value float64) {
	return
}

func (this *RediSession) Delete(key string) {
	return
}

func (this *RediSession) SetDeadline(second int) {
	return
}

func (this *RediSession) Deadline() int {
	return 0
}

func (this *RediSession) Die() {
	return
}

func (this *RediSession) Dead() bool {
	return false
}

type RediSessionContainer struct {
	sessionCounter int                     // session计数器
	sessions       map[string]*RediSession // 存储Session
	defaultExpire  int                     // 默认过期时间
	source         string                  // redis配置信息
	closed         bool                    // 是否关闭
	pool           *redis.Pool             // redis连接池
}

func NewRediSessionContainer(expire int, source string) (session.SessionContainer, error) {
	var container = new(RediSessionContainer)
	container.defaultExpire = expire
	container.source = source

	container.pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			var conn, err = redis.Dial("tcp", source)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	return container, nil
}

func (this *RediSessionContainer) CreateSession() (session.Session, bool) {
	if this.closed {
		return nil, false
	}
	var sessionId = util.NewUUID().Hex()
	var ss = newRediSession(sessionId)
	this.sessionCounter++
	this.sessions[sessionId] = ss
	return nil, false
}

// Session 获取Session,并且更新deadline,http processor 会在每次请求到的时候调用该方法
func (this *RediSessionContainer) Session(sessionId string) (session.Session, bool) {
	return nil, false
}

func (this *RediSessionContainer) Close() {
	this.closed = true
}

func (this *RediSessionContainer) Closed() bool {
	return this.closed
}
