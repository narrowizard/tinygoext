package sessionext

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"github.com/kdada/tinygo/session"
	"github.com/kdada/tinygo/util"
)

type sessionData struct {
	UpdatedAt time.Time
	Data      interface{}
}

type redisConfig struct {
	Host        string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int // seconds
	DB          int
	Password    string
	Wait        bool
}

type RediSession struct {
	sessionId string                 // 会话id
	data      map[string]sessionData // 数据
	dead      bool
}

func newRediSession(sessionId string) *RediSession {
	var ss = new(RediSession)
	ss.sessionId = sessionId
	ss.data = make(map[string]sessionData, 0)
	ss.dead = false
	return ss
}

func (this *RediSession) SessionId() string {
	return this.sessionId
}

func (this *RediSession) Value(key string) (interface{}, bool) {
	var v, ok = this.data[key]
	return v.Data, ok
}

func (this *RediSession) String(key string) (string, bool) {
	v, ok := this.Value(key)
	s, ok := v.(string)
	return s, ok
}

func (this *RediSession) Int(key string) (int, bool) {
	v, _ := this.Value(key)
	i, ok := v.(int)
	if !ok {
		s, ok := v.(float64)
		if ok {
			i = int(s)
			if float64(i) == s {
				return i, ok
			}
		}
		return 0, false
	}
	return i, ok
}

func (this *RediSession) Bool(key string) (bool, bool) {
	v, ok := this.Value(key)
	s, ok := v.(bool)
	return s, ok
}

func (this *RediSession) Float(key string) (float64, bool) {
	v, ok := this.Value(key)
	s, ok := v.(float64)
	return s, ok
}

func (this *RediSession) SetValue(key string, value interface{}) {
	var temp sessionData
	temp.Data = value
	temp.UpdatedAt = time.Now()
	this.data[key] = temp
}

func (this *RediSession) SetString(key string, value string) {
	this.SetValue(key, value)
}

func (this *RediSession) SetInt(key string, value int) {
	this.SetValue(key, value)
}

func (this *RediSession) SetBool(key string, value bool) {
	this.SetValue(key, value)
}

func (this *RediSession) SetFloat(key string, value float64) {
	this.SetValue(key, value)
}

func (this *RediSession) Delete(key string) {
	delete(this.data, key)
}

func (this *RediSession) SetDeadline(second int) {
	return
}

func (this *RediSession) Deadline() int {
	return 0
}

func (this *RediSession) Die() {
	this.dead = true
}

func (this *RediSession) Dead() bool {
	return this.dead
}

type RediSessionContainer struct {
	sessionCounter int                     // session计数器
	sessions       map[string]*RediSession // 存储Session
	defaultExpire  int                     // 默认过期时间
	rwm            sync.RWMutex            // 读写锁
	source         string                  // redis host
	closed         bool                    // 是否关闭
	pool           *redis.Pool             // redis连接池
	redlocks       *redsync.Redsync
}

func NewRediSessionContainer(expire int, source string) (session.SessionContainer, error) {
	var container = new(RediSessionContainer)
	container.defaultExpire = expire
	container.sessions = make(map[string]*RediSession, 100)
	container.closed = false
	// 解析配置
	var config redisConfig
	var err = json.Unmarshal([]byte(source), &config)
	if err != nil {
		panic(err)
	}
	container.source = config.Host
	container.pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			var conn, err = redis.Dial("tcp", container.source)
			if err != nil {
				return nil, err
			}
			if config.Password != "" {
				if _, err := conn.Do("AUTH", config.Password); err != nil {
					conn.Close()
					return nil, err
				}
			}
			if _, err := conn.Do("SELECT", config.DB); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		},
		MaxActive:   config.MaxActive,
		MaxIdle:     config.MaxIdle,
		IdleTimeout: time.Second * time.Duration(config.IdleTimeout),
		Wait:        config.Wait,
	}
	var pools []redsync.Pool
	pools = append(pools, container.pool)
	container.redlocks = redsync.New(pools)
	return container, nil
}

func (this *RediSessionContainer) CreateSession() (session.Session, bool) {
	if this.closed {
		return nil, false
	}
	this.rwm.Lock()
	defer this.rwm.Unlock()
	var sessionId = util.NewUUID().Hex()
	var ss = newRediSession(sessionId)
	this.sessionCounter++
	this.sessions[sessionId] = ss
	// write to redis
	this.writeRedis(ss)
	return ss, true
}

// Session 获取Session,并且更新deadline,http processor 会在每次请求到的时候调用该方法
func (this *RediSessionContainer) Session(sessionId string) (session.Session, bool) {
	if this.closed {
		return nil, false
	}
	// sync data to redis
	var ss, ok = this.syncRedis(sessionId)
	return ss, ok
}

func (this *RediSessionContainer) Close() {
	this.closed = true
}

func (this *RediSessionContainer) Closed() bool {
	return this.closed
}

// writeRedis 将session数据写到redis中
func (this *RediSessionContainer) writeRedis(session *RediSession) bool {
	var conn = this.pool.Get()
	defer conn.Close()
	var bData, err = json.Marshal(session.data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	var lock = this.redlocks.NewMutex(session.sessionId + "_lock")
	// 加锁
	err = lock.Lock()
	if err != nil {
		return false
	}
	conn.Send("MULTI")
	conn.Send("Set", session.sessionId, string(bData))
	conn.Send("Expire", session.sessionId, this.defaultExpire)
	_, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		return false
	}
	// 解锁
	lock.Unlock()
	return true
}

// syncRedis 与redis同步
func (this *RediSessionContainer) syncRedis(sessionId string) (*RediSession, bool) {
	this.rwm.RLock()
	var session, ok = this.sessions[sessionId]
	this.rwm.RUnlock()
	if !ok {
		// session 在本地不存在, create session
		session = newRediSession(sessionId)
		this.rwm.Lock()
		this.sessions[sessionId] = session
		this.rwm.Unlock()
	}
	var conn = this.pool.Get()
	defer conn.Close()
	if session.Dead() {
		// 从远程删除
		conn.Do("Del", sessionId)
		// 从本地删除
		this.rwm.Lock()
		delete(this.sessions, sessionId)
		this.rwm.Unlock()
		return nil, false
	}
	var data, err = redis.String(conn.Do("Get", sessionId))
	if err != nil {
		// session 在redis中不存在, 删除
		this.rwm.Lock()
		delete(this.sessions, sessionId)
		this.rwm.Unlock()
		return nil, false
	}
	var mData = make(map[string]sessionData)
	err = json.Unmarshal([]byte(data), &mData)
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	for k, v := range mData {
		var localValue, ok = session.data[k]
		if !ok || localValue.UpdatedAt.Before(v.UpdatedAt) {
			// 本地不存在 或 远程数据更新
			session.data[k] = v
		}
	}
	return session, this.writeRedis(session)
}
