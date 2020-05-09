package session

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

//cookie内存放的sessionId 键名
var (
	cookieSessionName = "GO_WEBSESS"
)

//CookieOptions cookie存放的基础属性
//路径、所属域、存活时间、是否安全、只经由http传输
type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
}

//Session 结构
//id、值、cookie属性、是否是新会话、最后活跃时间、仓库
type Session struct {
	ID      string
	Values  map[interface{}]interface{}
	Options *CookieOptions
	storage IStorage
	IsNew   bool
	ActTime int64
}

//Set 设置session值
func (s *Session) Set(key interface{}, value interface{}) {
	s.Values[key] = value
}

//Get 获取session内的值
//获取后需要自行断言
func (s *Session) Get(key interface{}) interface{} {
	var value interface{}
	if value, ok := s.Values[key]; ok {
		return value
	}
	return value
}

//Del 删除某个session值
func (s *Session) Del(key interface{}) {
	if _, ok := s.Values[key]; ok {
		delete(s.Values, key)
	}
}

//Len 获取一个session中值的个数
func (s *Session) Len() (n int) {
	n = len(s.Values)
	return
}

//Save 将session保存
func (s *Session) Save(w http.ResponseWriter, r *http.Request) {
	s.ActTime = time.Now().Unix()
	s.storage.Save(w, r, s)
}

//GC session 垃圾回收判断
func (s *Session) GC() bool {
	return int(time.Now().Unix()-s.ActTime) > s.Options.MaxAge
}

//NewCookie 生成一个新的Cookie结构
func NewCookie(s *Session) *http.Cookie {
	cookie := &http.Cookie{
		Name:   cookieSessionName,
		Value:  s.ID,
		Path:   s.Options.Path,
		Domain: s.Options.Domain,
		Secure: s.Options.Secure,
		MaxAge: s.Options.MaxAge,
	}
	return cookie
}

//随机数因子
//用以解决windows下出现的同一时刻
//会产生同一随机数的问题
var randSeed int64 = 0

//生成随机sessionID
func createSessionID() string {
	rand.Seed(time.Now().UnixNano())
	var result bytes.Buffer
	for i := 0; i < 10; {
		c := getChar()
		result.WriteByte(c)
		i++
	}
	randSeed++
	return result.String()
}

//获取随机字符串
func getChar() byte {
	var c int
	switch rand.Intn(3) {
	case 0:
		c = 65 + rand.Intn(90-65)
	case 1:
		c = 97 + rand.Intn(122-97)
	default:
		c = 48 + rand.Intn(9)
	}
	return byte(c)
}

//Init 初始化引擎
func Init(store IStorage, cookieName ...string) {
	if len(cookieName) > 0 {
		cookieSessionName = cookieName[0]
	}
	storage = store
	storage.GC()
}

//SetCookieSessionName 设置cookieSessionName 来取代默认值
func SetCookieSessionName(s string) {
	cookieSessionName = s
}

//主要用于将Session保存到字符串的中间结构
type serializableSession struct {
	ID      string
	Values  map[string]interface{}
	Options *CookieOptions
	storage IStorage
	IsNew   bool
	ActTime int64
}

//序列化session数据
func serializeSession(sess *Session) ([]byte, error) {
	s := &serializableSession{}
	s.ID = sess.ID
	s.IsNew = sess.IsNew
	s.Options = sess.Options
	s.ActTime = sess.ActTime
	s.Values = transSaveValueType(sess.Values)

	return json.Marshal(s)
}

//反序列化session数据
func unserializeSession(data []byte) (*Session, error) {
	session := &serializableSession{}
	err := json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}
	return &Session{
		storage: storage,
		ID:      session.ID,
		ActTime: session.ActTime,
		Options: session.Options,
		IsNew:   session.IsNew,
		Values:  transGetValueType(session.Values),
	}, nil
}

//将map[interface{}]interface{} -> map[string]interface{}
func transSaveValueType(i map[interface{}]interface{}) (s map[string]interface{}) {
	s = make(map[string]interface{}, 10)
	for k, v := range i {
		if val, ok := k.(string); ok {
			s[val] = v
		}
	}
	return
}

//将map[string]interface{} -> map[interface{}]interface{}
func transGetValueType(s map[string]interface{}) (i map[interface{}]interface{}) {
	i = make(map[interface{}]interface{})
	for k, v := range s {
		i[k] = v
	}
	return
}
