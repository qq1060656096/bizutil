package maputil

import (
	"testing"
)

// ============== MapGet 测试 ==============

func TestMapGet_KeyExists(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	v, ok := MapGet(m, "b", func(i int) int { return i * 10 })
	if !ok {
		t.Error("expected ok to be true")
	}
	if v != 20 {
		t.Errorf("expected v to be 20, got %d", v)
	}
}

func TestMapGet_KeyNotExists(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	v, ok := MapGet(m, "notexist", func(i int) int { return i * 10 })
	if ok {
		t.Error("expected ok to be false")
	}
	if v != 0 {
		t.Errorf("expected v to be zero value (0), got %d", v)
	}
}

func TestMapGet_ValueFuncNil(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	v, ok := MapGet[int, string, int](m, "a", nil)
	if !ok {
		t.Error("expected ok to be true when key exists")
	}
	if v != 0 {
		t.Errorf("expected v to be zero value (0) when value func is nil, got %d", v)
	}
}

func TestMapGet_ValueFuncNil_KeyNotExists(t *testing.T) {
	m := map[string]int{"a": 1}
	v, ok := MapGet[int, string, int](m, "notexist", nil)
	if ok {
		t.Error("expected ok to be false when key not exists")
	}
	if v != 0 {
		t.Errorf("expected v to be zero value (0), got %d", v)
	}
}

func TestMapGet_EmptyMap(t *testing.T) {
	m := map[string]int{}
	v, ok := MapGet(m, "any", func(i int) int { return i })
	if ok {
		t.Error("expected ok to be false for empty map")
	}
	if v != 0 {
		t.Errorf("expected v to be zero value (0), got %d", v)
	}
}

func TestMapGet_NilMap(t *testing.T) {
	var m map[string]int
	v, ok := MapGet(m, "any", func(i int) int { return i })
	if ok {
		t.Error("expected ok to be false for nil map")
	}
	if v != 0 {
		t.Errorf("expected v to be zero value (0), got %d", v)
	}
}

func TestMapGet_DifferentKeyType_Int(t *testing.T) {
	m := map[int]string{1: "one", 2: "two"}
	v, ok := MapGet(m, 1, func(s string) string { return s + "!" })
	if !ok {
		t.Error("expected ok to be true")
	}
	if v != "one!" {
		t.Errorf("expected v to be 'one!', got %s", v)
	}
}

func TestMapGet_TransformType(t *testing.T) {
	// 测试 value 函数返回不同类型
	m := map[string]int{"count": 42}
	v, ok := MapGet(m, "count", func(i int) string {
		if i > 10 {
			return "large"
		}
		return "small"
	})
	if !ok {
		t.Error("expected ok to be true")
	}
	if v != "large" {
		t.Errorf("expected v to be 'large', got %s", v)
	}
}

func TestMapGet_StructValue(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	m := map[int]User{
		1: {Name: "Alice", Age: 30},
		2: {Name: "Bob", Age: 25},
	}
	v, ok := MapGet(m, 1, func(u User) string { return u.Name })
	if !ok {
		t.Error("expected ok to be true")
	}
	if v != "Alice" {
		t.Errorf("expected v to be 'Alice', got %s", v)
	}
}

func TestMapGet_PointerValue(t *testing.T) {
	type Data struct {
		Value int
	}
	m := map[string]*Data{
		"x": {Value: 100},
	}
	v, ok := MapGet(m, "x", func(d *Data) int { return d.Value })
	if !ok {
		t.Error("expected ok to be true")
	}
	if v != 100 {
		t.Errorf("expected v to be 100, got %d", v)
	}
}

func TestMapGet_NilPointerInMap(t *testing.T) {
	m := map[string]*int{
		"nil": nil,
	}
	// 当 value 为 nil 指针时，value 函数仍会被调用
	v, ok := MapGet(m, "nil", func(i *int) bool { return i == nil })
	if !ok {
		t.Error("expected ok to be true")
	}
	if !v {
		t.Error("expected v to be true (nil pointer)")
	}
}

// ============== MapBy 测试 ==============

func TestMapBy_Basic(t *testing.T) {
	list := []string{"apple", "banana", "cherry"}
	m := MapBy(list, func(s string) string { return s[:1] }, func(s string) int { return len(s) })

	if len(m) != 3 {
		t.Errorf("expected map length 3, got %d", len(m))
	}
	if m["a"] != 5 {
		t.Errorf("expected m['a'] = 5, got %d", m["a"])
	}
	if m["b"] != 6 {
		t.Errorf("expected m['b'] = 6, got %d", m["b"])
	}
	if m["c"] != 6 {
		t.Errorf("expected m['c'] = 6, got %d", m["c"])
	}
}

func TestMapBy_EmptySlice(t *testing.T) {
	list := []int{}
	m := MapBy(list, func(i int) int { return i }, func(i int) string { return "x" })
	if len(m) != 0 {
		t.Errorf("expected empty map, got length %d", len(m))
	}
}

func TestMapBy_NilSlice(t *testing.T) {
	var list []int
	m := MapBy(list, func(i int) int { return i }, func(i int) string { return "x" })
	if m == nil {
		t.Error("expected non-nil map")
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got length %d", len(m))
	}
}

func TestMapBy_DuplicateKeys_LastWins(t *testing.T) {
	// 测试相同 key 时后面的值覆盖前面的
	type Item struct {
		ID   int
		Name string
	}
	list := []Item{
		{ID: 1, Name: "first"},
		{ID: 2, Name: "second"},
		{ID: 1, Name: "third"}, // 重复 ID
	}
	m := MapBy(list, func(i Item) int { return i.ID }, func(i Item) string { return i.Name })

	if len(m) != 2 {
		t.Errorf("expected map length 2, got %d", len(m))
	}
	if m[1] != "third" {
		t.Errorf("expected m[1] = 'third' (last wins), got %s", m[1])
	}
	if m[2] != "second" {
		t.Errorf("expected m[2] = 'second', got %s", m[2])
	}
}

func TestMapBy_StructToMap(t *testing.T) {
	type User struct {
		ID    int
		Name  string
		Email string
	}
	users := []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}
	m := MapBy(users, func(u User) int { return u.ID }, func(u User) User { return u })

	if len(m) != 2 {
		t.Errorf("expected map length 2, got %d", len(m))
	}
	if m[1].Name != "Alice" {
		t.Errorf("expected m[1].Name = 'Alice', got %s", m[1].Name)
	}
}

func TestMapBy_IdentityValue(t *testing.T) {
	// 测试 value 函数直接返回原值
	list := []int{10, 20, 30}
	m := MapBy(list, func(i int) int { return i / 10 }, func(i int) int { return i })

	if m[1] != 10 {
		t.Errorf("expected m[1] = 10, got %d", m[1])
	}
	if m[2] != 20 {
		t.Errorf("expected m[2] = 20, got %d", m[2])
	}
	if m[3] != 30 {
		t.Errorf("expected m[3] = 30, got %d", m[3])
	}
}

func TestMapBy_StringKey(t *testing.T) {
	type Product struct {
		SKU   string
		Price float64
	}
	products := []Product{
		{SKU: "A001", Price: 9.99},
		{SKU: "B002", Price: 19.99},
	}
	m := MapBy(products, func(p Product) string { return p.SKU }, func(p Product) float64 { return p.Price })

	if m["A001"] != 9.99 {
		t.Errorf("expected m['A001'] = 9.99, got %f", m["A001"])
	}
	if m["B002"] != 19.99 {
		t.Errorf("expected m['B002'] = 19.99, got %f", m["B002"])
	}
}

func TestMapBy_PointerElements(t *testing.T) {
	type Data struct {
		Key   string
		Value int
	}
	list := []*Data{
		{Key: "x", Value: 1},
		{Key: "y", Value: 2},
	}
	m := MapBy(list, func(d *Data) string { return d.Key }, func(d *Data) int { return d.Value })

	if m["x"] != 1 {
		t.Errorf("expected m['x'] = 1, got %d", m["x"])
	}
	if m["y"] != 2 {
		t.Errorf("expected m['y'] = 2, got %d", m["y"])
	}
}

func TestMapBy_SingleElement(t *testing.T) {
	list := []int{42}
	m := MapBy(list, func(i int) string { return "key" }, func(i int) int { return i * 2 })

	if len(m) != 1 {
		t.Errorf("expected map length 1, got %d", len(m))
	}
	if m["key"] != 84 {
		t.Errorf("expected m['key'] = 84, got %d", m["key"])
	}
}

func TestMapBy_ComplexTransform(t *testing.T) {
	// 复杂转换：将字符串切片转为 map[首字母大写]长度
	list := []string{"apple", "apricot", "banana"}
	m := MapBy(list,
		func(s string) byte { return s[0] },
		func(s string) int { return len(s) },
	)

	// "apricot" 会覆盖 "apple" (都是 'a' 开头)
	if m['a'] != 7 {
		t.Errorf("expected m['a'] = 7 (apricot length), got %d", m['a'])
	}
	if m['b'] != 6 {
		t.Errorf("expected m['b'] = 6, got %d", m['b'])
	}
}

func TestMapBy_AllSameKey(t *testing.T) {
	// 所有元素生成相同的 key
	list := []int{1, 2, 3, 4, 5}
	m := MapBy(list, func(i int) string { return "same" }, func(i int) int { return i })

	if len(m) != 1 {
		t.Errorf("expected map length 1, got %d", len(m))
	}
	if m["same"] != 5 {
		t.Errorf("expected m['same'] = 5 (last element), got %d", m["same"])
	}
}
