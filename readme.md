### session存储工具

[历史版本源码位置](https://github.com/kasiss-liu/go-tools/tree/master/sessions)

#### Usage

```go
# store   IStorage
session.Init(store)

sess := NewSession("/", "localhost", 300, true, false)

sess.Set("username", "foo")
sess.Set("testing", "bar")

length := sess.Len()
val := sess.Get("testing")
sess.Del("testing")

fmt.Println(length)
fmt.Println(val)
fmt.Println(sess.Get("testing"))
ftm.Printf("%#v\n",sess.Values)
/**
    2
    bar
    interface{}
    map[interface{}]interface{}{"username":"foo"}
*/

```
