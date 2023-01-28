//go:build go1.9
// +build go1.9

package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"
)

//func init() {
//    //etcdAddress = "http://172.16.1.248:2379"
//    etcdAddress = "http://127.0.0.1:2379"
//    GetOrNewMockTestEtcd(etcdAddress)
//    if !etcd_mock_started {
//        go testEtcd.StartMockEtcd()
//    }
//    testEtcd.WaitingForEtcdStarted()
//}

func TestEtcdV3KeyValueConfigSource(t *testing.T) {

	address := testEtcd.Address

	root := "/config101/test/kvdemo1"
	size := 10
	m := initEtcdV3Data(address, root, size)
	c := NewEtcdV3KeyValueConfigSource(address, root)
	Convey("etcd kv api3", t, func() {
		keys := c.Keys()
		So(len(keys), ShouldEqual, size)
		for _, key := range keys {
			v, ok := m[key]
			//fmt.Println(key)
			v1, err := c.Get(key)
			So(ok, ShouldEqual, true)
			So(err, ShouldBeNil)
			So(v, ShouldEqual, v1)
		}
	})

}

func initEtcdV3Data(address, root string, size int) map[string]string {

	cfg := clientv3.Config{
		Endpoints:   []string{address},
		DialTimeout: 3 * time.Second,
	}
	c, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	//kv := clientv3.NewKV(c)
	kv := namespace.NewKV(c, root)
	r, e := kv.Delete(context.Background(), "/", clientv3.WithPrefix())
	fmt.Println(r, e)
	m := make(map[string]string)
	for i := 0; i < size; i++ {
		key := "/key/x" + strconv.Itoa(i)
		//keyFull := filepath.Join(root, key)
		value := "value-" + strconv.Itoa(i)
		kv.Put(context.Background(), key, value)
		//fmt.Println(res, err)
		k := strings.Replace(key, "/", ".", -1)
		//fmt.Println(key, k, value)
		m[k] = value
	}
	//c.KV = namespace.NewKV(c.KV, root)
	res, err := kv.Get(context.Background(), "/", clientv3.WithPrefix(), clientv3.WithKeysOnly())
	//res, err := c.KV.Get(context.Background(), "key", clientv3.WithKeysOnly())
	fmt.Println(res.Kvs, res.Count, res.More)
	return m

}
