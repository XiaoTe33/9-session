package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var mgr = NewSessionMgr("my_session", 10)

func main() {
	r := gin.Default()

	r.GET("/session/1", SessionMiddleware, func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "哦耶"})
	})

	r.Run(":8081")
}

type SessionMgr struct {
	cookieName  string
	maxLifeTime int64
	sessions    map[string]*Session
}
type Session struct {
	sessionID string
	lastTime  time.Time
	values    map[interface{}]interface{} //?这是咋存储的
}

func SessionMiddleware(c *gin.Context) {
	//1.Read Cookie
	sessionID, err := mgr.ReadCookie(c)
	if err != nil {
		sessionID = mgr.NewSession(c)
		c.SetCookie(mgr.cookieName, sessionID, int(mgr.maxLifeTime), "/", "localhost", false, true)
		c.Next()
		return
	}

	//2.Check ID
	err = mgr.CheckID(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}
	c.Next()

}

// CheckID 检查是否存在
func (m *SessionMgr) CheckID(sessionID string) error {
	session, ok := mgr.sessions[sessionID]
	if ok {
		session.lastTime = time.Now()
		return nil
	}
	return errors.New("sessionID is invalid")
}

// ReadCookie 读取Cookie
func (m *SessionMgr) ReadCookie(c *gin.Context) (sessionID string, err error) {
	sessionID, err = c.Cookie(mgr.cookieName)
	if err != nil {
		fmt.Println("cookie not set")
	}
	return sessionID, err
}

// NewSession 创建一个Session并且存到map里
func (m *SessionMgr) NewSession(c *gin.Context) string {
	newSessionID := url.QueryEscape(mgr.NewSessionID())
	session := &Session{sessionID: newSessionID, lastTime: time.Now(),
		values: make(map[interface{}]interface{})} //创建session，设置sessionID和最后时间
	session.values = map[interface{}]interface{}{"k": "v"}
	//这里应该还要根据请求设置values的信息？

	mgr.sessions[newSessionID] = session //把session存在服务器里
	return newSessionID

}

// NewSessionMgr Creat Manager来控制sessions
func NewSessionMgr(cookieName string, maxLifeTime int64) *SessionMgr {
	mgr := &SessionMgr{cookieName: cookieName, maxLifeTime: maxLifeTime, sessions: make(map[string]*Session)}
	go mgr.SessionGC()
	return mgr
}

// SessionGC 时间到了就清扫一次
func (m *SessionMgr) SessionGC() {
	for id, session := range m.sessions {
		if session.lastTime.Unix()+m.maxLifeTime < time.Now().Unix() {
			delete(m.sessions, id)
		}
	}
	time.AfterFunc(time.Duration(m.maxLifeTime)*time.Second, func() {
		m.SessionGC()
	})
}

// NewSessionID 生成sessionID
func (m *SessionMgr) NewSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		nano := time.Now().UnixNano()
		return strconv.FormatInt(nano, 10)
	}
	return base64.URLEncoding.EncodeToString(b)
}
