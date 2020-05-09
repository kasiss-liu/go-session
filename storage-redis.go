package session

import (
	"errors"
	"net/http"

	"github.com/gomodule/redigo/redis"
)

//RedisStorage 基于redis的session仓库
type RedisStorage struct {
	network string
	address string
	conn    redis.Conn
}

//Save 存储session
func (rs *RedisStorage) Save(w http.ResponseWriter, r *http.Request, sess *Session) error {

	data, err := serializeSession(sess)
	if err != nil {
		return err
	}
	err = rs.conn.Send("SETEX", sess.ID, sess.Options.MaxAge, string(data))
	rs.conn.Flush()
	if err != nil {
		return err
	}
	_, err = rs.conn.Receive()
	if err != nil {
		return err
	}
	return nil
}

//Get 获取session数据
//
func (rs *RedisStorage) Get(r *http.Request, name string) (sess *Session, err error) {
	err = rs.conn.Send("GET", name)
	rs.conn.Flush()
	if err != nil {
		return nil, err
	}
	v, err := rs.conn.Receive()
	if err != nil {
		return nil, err
	}
	if bs, ok := v.([]byte); ok {
		sess, err := unserializeSession(bs)
		return sess, err
	}
	return nil, errors.New("data is not available")
}

//Del 按key删除session
func (rs *RedisStorage) Del(name string) {
	rs.conn.Send("del", name)
	rs.conn.Flush()
}

//GC 基于redis过期机制 GC逻辑可忽略
func (rs *RedisStorage) GC() {}

//NewRedisSessionStorage 基于redis链接构建一个session仓库
func NewRedisSessionStorage(network, address string) (*RedisStorage, error) {
	conn, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	rs := &RedisStorage{
		network: network,
		address: address,
		conn:    conn,
	}
	return rs, nil
}
