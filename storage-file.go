package session

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//FileStorage 文件session仓库
type FileStorage struct {
	storagePath string
	prefix      string
	list        map[string]string
	rwLock      sync.RWMutex
}

//Save 保存session
//向http请求中写入数据并保存到session内容至文件
func (fs *FileStorage) Save(w http.ResponseWriter, r *http.Request, sess *Session) error {
	fs.rwLock.Lock()
	defer fs.rwLock.Unlock()
	//处理session的名称和文件名
	name := sess.ID
	filename := fs.prefix + name
	//将*Session 序列化为json []byte数据
	data, err := serializeSession(sess)
	if err != nil {
		return err
	}
	//将session内容写入文件
	err = fs.writeSessionFile(filename, data)
	if err != nil {
		return err
	}
	//写入http请求
	if sess.IsNew {
		sess.IsNew = false
		fs.list[name] = filename
		http.SetCookie(w, NewCookie(sess))
	}
	return nil
}

//Get 从http请求的cookie中获取sessionID 并读取相应的session文件
//返回*Session
func (fs *FileStorage) Get(r *http.Request, name string) (*Session, error) {
	fs.rwLock.RLock()
	defer fs.rwLock.RUnlock()
	//从map中获取session文件名 然后读取session文件
	if sessName, ok := fs.list[name]; ok {
		data, err := fs.readSessionFile(sessName)
		if err != nil {
			return nil, err
		}
		//将获取到的字符串重新转为一个*Session 并返回
		return unserializeSession(data)
	}
	return nil, errors.New("session lost")
}

//Del 删除session 并将对应的session文件删除
func (fs *FileStorage) Del(name string) {
	fs.rwLock.Lock()
	defer fs.rwLock.Unlock()
	if filename, ok := fs.list[name]; ok {
		delete(fs.list, name)
		os.Remove(filename)
	}
}

//GC session回收
//每秒轮询list内的session数据
//如果文件内容损坏、
//如果文件丢失、
//如果session超时
//将filename从list中移除 并删除文件
func (fs *FileStorage) GC() {
	go func() {
		for {
			for name, filename := range fs.list {
				data, err := fs.readSessionFile(filename)
				if err != nil && filename != "" {
					fs.Del(name)
					os.Remove(fs.storagePath + filename)
					continue
				}
				sess, err := unserializeSession(data)
				if err != nil {
					fs.Del(name)
					os.Remove(fs.storagePath + filename)
					continue
				}
				if sess.GC() {
					fs.Del(name)
					os.Remove(fs.storagePath + filename)
					continue
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

//从session文件中读取内容
//并返回字符串和途中遇到的error
func (fs *FileStorage) readSessionFile(name string) ([]byte, error) {
	file, err := os.Open(fs.storagePath + name)
	defer file.Close()
	if err != nil {
		return []byte{}, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

//向session文件内写入session内容
func (fs *FileStorage) writeSessionFile(name string, data []byte) error {
	filename := fs.storagePath + name

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		_, err := os.Create(filename)

		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.WriteString(file, string(data))
	return err
}

//NewFileSessionStorage 初始化一个文件session仓库
//判断存储路径是否可用（是否存在、是否可写）
//设置文件存储路径
//设置session文件的前缀prefix
//将未清理的session初始化至内存中 继续使用
func NewFileSessionStorage(path string, prefix ...string) IStorage {
	var sessionPrefix string
	var err error
	//判断路径是否可写
	if runtime.GOOS != "windows" {
		_, err = os.Stat(path)
		if os.IsNotExist(err) || os.IsPermission(err) {
			panic(err.Error())
		}
	}

	//获取定义的session前缀（可选）
	if len(prefix) > 0 {
		sessionPrefix = prefix[0]
	} else {
		sessionPrefix = "sess_"
	}
	//判断路径是否可用
	file, err := os.Stat(path)
	if err != nil {
		panic(err.Error())
	}
	//判断路径是否为文件夹
	if !file.IsDir() {
		panic("session store path is not directory")
	}
	//在传递的路径右侧添加unix 路径分隔符
	path = strings.TrimRight(path, "/") + "/"
	//将未清理的session文件初始化到内存中
	list := make(map[string]string, 100)
	//遍历存储路径下的文件
	filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		//判断文件是否是session文件
		filename := filepath.Base(p)
		//如果没有session前缀 则丢弃
		if !strings.HasPrefix(filename, sessionPrefix) {
			return nil
		}
		//获取name
		name := strings.Replace(filename, sessionPrefix, "", -1)
		list[name] = filename
		return nil
	})
	//生成一个新的session仓库
	return &FileStorage{storagePath: path, prefix: sessionPrefix, list: list}
}
