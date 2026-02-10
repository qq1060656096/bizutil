package sliceutil

import (
	"errors"
	"testing"
)

type Person struct {
	ID   int
	Name string
}

func TestFindDuplicateWithErr(t *testing.T) {
	tests := []struct {
		name    string
		list    []Person
		keyFn   func(Person) int
		errFn   func(Person) error
		wantErr bool
		errMsg  string
	}{
		{
			name: "无重复元素",
			list: []Person{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 3, Name: "Charlie"},
			},
			keyFn:   func(p Person) int { return p.ID },
			errFn:   nil,
			wantErr: false,
		},
		{
			name: "有重复元素，使用默认错误",
			list: []Person{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 1, Name: "Alice Duplicate"},
			},
			keyFn:   func(p Person) int { return p.ID },
			errFn:   nil,
			wantErr: true,
			errMsg:  "重复值: {ID:1 Name:Alice Duplicate}",
		},
		{
			name: "有重复元素，使用自定义错误函数",
			list: []Person{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 1, Name: "Alice Duplicate"},
			},
			keyFn: func(p Person) int { return p.ID },
			errFn: func(p Person) error {
				return errors.New("自定义错误: ID 重复")
			},
			wantErr: true,
			errMsg:  "自定义错误: ID 重复",
		},
		{
			name:    "空列表",
			list:    []Person{},
			keyFn:   func(p Person) int { return p.ID },
			errFn:   nil,
			wantErr: false,
		},
		{
			name:    "单个元素",
			list:    []Person{{ID: 1, Name: "Alice"}},
			keyFn:   func(p Person) int { return p.ID },
			errFn:   nil,
			wantErr: false,
		},
		{
			name: "多个相同元素，第一个重复",
			list: []Person{
				{ID: 1, Name: "Alice"},
				{ID: 1, Name: "Alice 2"},
				{ID: 1, Name: "Alice 3"},
			},
			keyFn: func(p Person) int { return p.ID },
			errFn: func(p Person) error {
				return errors.New("发现重复: " + p.Name)
			},
			wantErr: true,
			errMsg:  "发现重复: Alice 2",
		},
		{
			name: "多个相同元素，第一个重复",
			list: []Person{
				{ID: 1, Name: "Alice"},
				{ID: 1, Name: "Alice 2"},
				{ID: 1, Name: "Alice 3"},
			},
			keyFn: func(p Person) int { return p.ID },
			errFn: func(p Person) error {
				return errors.New("发现重复: " + p.Name)
			},
			wantErr: true,
			errMsg:  "发现重复: Alice 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FindDuplicateWithErr[Person, int](tt.list, tt.keyFn, tt.errFn)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindDuplicateWithErr() 期望返回错误，但没有返回")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("FindDuplicateWithErr() 错误消息不匹配，期望: %s, 实际: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("FindDuplicateWithErr() 不期望返回错误，但返回了: %v", err)
				}
			}
		})
	}
}

func TestFindDuplicateWithErr_WithStringKey(t *testing.T) {
	tests := []struct {
		name    string
		list    []Person
		wantErr bool
		errMsg  string
	}{
		{
			name: "使用字符串作为键无重复",
			list: []Person{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 3, Name: "Charlie"},
			},
			wantErr: false,
		},
		{
			name: "使用字符串作为键有重复",
			list: []Person{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 3, Name: "Alice"},
			},
			wantErr: true,
			errMsg:  "姓名重复: Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FindDuplicateWithErr(tt.list,
				func(p Person) string { return p.Name },
				func(p Person) error {
					return errors.New("姓名重复: " + p.Name)
				})

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindDuplicateWithErr() 期望返回错误，但没有返回")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("FindDuplicateWithErr() 错误消息不匹配，期望: %s, 实际: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("FindDuplicateWithErr() 不期望返回错误，但返回了: %v", err)
				}
			}
		})
	}
}

func TestFindDuplicateWithErr_WithIntegers(t *testing.T) {
	tests := []struct {
		name    string
		list    []int
		wantErr bool
	}{
		{
			name:    "整数无重复",
			list:    []int{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "整数有重复",
			list:    []int{1, 2, 3, 2, 5},
			wantErr: true,
		},
		{
			name:    "整数空列表",
			list:    []int{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FindDuplicateWithErr(tt.list, func(i int) int { return i }, nil)
			if tt.wantErr && err == nil {
				t.Errorf("FindDuplicateWithErr() 期望返回错误，但没有返回")
			} else if !tt.wantErr && err != nil {
				t.Errorf("FindDuplicateWithErr() 不期望返回错误，但返回了: %v", err)
			}
		})
	}
}

func TestFindDuplicateWithErr_WithStrings(t *testing.T) {
	tests := []struct {
		name    string
		list    []string
		wantErr bool
	}{
		{
			name:    "字符串无重复",
			list:    []string{"apple", "banana", "cherry"},
			wantErr: false,
		},
		{
			name:    "字符串有重复",
			list:    []string{"apple", "banana", "apple"},
			wantErr: true,
		},
		{
			name:    "字符串空列表",
			list:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FindDuplicateWithErr(tt.list, func(s string) string { return s }, nil)
			if tt.wantErr && err == nil {
				t.Errorf("FindDuplicateWithErr() 期望返回错误，但没有返回")
			} else if !tt.wantErr && err != nil {
				t.Errorf("FindDuplicateWithErr() 不期望返回错误，但返回了: %v", err)
			}
		})
	}
}

// 基准测试
func BenchmarkFindDuplicateWithErr(b *testing.B) {
	list := make([]Person, 1000)
	for i := 0; i < 1000; i++ {
		list[i] = Person{ID: i, Name: "Person" + string(rune(i))}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindDuplicateWithErr(list, func(p Person) int { return p.ID }, nil)
	}
}

func BenchmarkFindDuplicateWithErr_WithDuplicates(b *testing.B) {
	list := make([]Person, 1000)
	for i := 0; i < 999; i++ {
		list[i] = Person{ID: i, Name: "Person" + string(rune(i))}
	}
	list[999] = Person{ID: 500, Name: "Duplicate"} // 创建一个重复

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindDuplicateWithErr(list, func(p Person) int { return p.ID }, nil)
	}
}
