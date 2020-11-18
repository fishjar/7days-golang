package lru

import (
	"reflect"
	"testing"
)

type String string // 自定义类型

func (d String) Len() int { // 实现Value接口
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)                                          // 创建实例
	lru.Add("key1", String("1234"))                                    // 添加缓存数据
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" { // 判断能否获取数据，以及数据对不对
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok { // 判断能否获取不存在的key的数据
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2) // 不包含k3的数据
	lru := New(int64(cap), nil)   // 容量限制不允许所有数据
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3)) // 添加后，k1应该会从缓存中移除

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 { // 如果k1还存在，或者双链表长度不是2
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456")) // 将会删除
	lru.Add("k2", String("k2"))       // 将会删除
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) { // 比较字符串切片
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestAdd(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key", String("1"))   // 添加缓存
	lru.Add("key", String("111")) // 顶替前面添加的缓存

	if lru.nbytes != int64(len("key")+len("111")) { // 判断缓存大小是否维护正确
		t.Fatal("expected 6 but got", lru.nbytes)
	}
}
