package tinygoRedisSession

import "github.com/kdada/tinygo/session"

type RediSession struct {
	sessionId string
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
	defaultExpire int    // 默认过期时间
	source        string // host地址
	closed        bool   //是否关闭
}

func NewRediSessionContainer(expire int, source string) (session.SessionContainer, error) {
	var container = new(RediSessionContainer)
	container.defaultExpire = expire
	container.source = source
	return container, nil
}

func (this *RediSessionContainer) CreateSession() (session.Session, bool) {
	return nil, false
}

func (this *RediSessionContainer) Session(sessionId string) (session.Session, bool) {
	return nil, false
}

func (this *RediSessionContainer) Close() {
	this.closed = true
}

func (this *RediSessionContainer) Closed() bool {
	return this.closed
}
