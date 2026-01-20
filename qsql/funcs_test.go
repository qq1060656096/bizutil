package qsql

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestGetValueByPath(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		paths     []string
		wantValue interface{}
		wantOk    bool
	}{
		{
			name:      "单层路径-字符串",
			jsonData:  `{"name": "张三"}`,
			paths:     []string{"name"},
			wantValue: "张三",
			wantOk:    true,
		},
		{
			name:      "单层路径-数字",
			jsonData:  `{"age": 25}`,
			paths:     []string{"age"},
			wantValue: float64(25),
			wantOk:    true,
		},
		{
			name:      "单层路径-布尔值",
			jsonData:  `{"active": true}`,
			paths:     []string{"active"},
			wantValue: true,
			wantOk:    true,
		},
		{
			name:      "多层路径-单参数",
			jsonData:  `{"user": {"name": "李四"}}`,
			paths:     []string{"user.name"},
			wantValue: "李四",
			wantOk:    true,
		},
		{
			name:      "多层路径-多参数拼接",
			jsonData:  `{"user": {"profile": {"email": "test@example.com"}}}`,
			paths:     []string{"user", "profile", "email"},
			wantValue: "test@example.com",
			wantOk:    true,
		},
		{
			name:      "路径不存在",
			jsonData:  `{"name": "张三"}`,
			paths:     []string{"age"},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name:      "嵌套路径不存在",
			jsonData:  `{"user": {"name": "张三"}}`,
			paths:     []string{"user", "email"},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name:      "空JSON对象",
			jsonData:  `{}`,
			paths:     []string{"name"},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name:      "数组访问",
			jsonData:  `{"items": [1, 2, 3]}`,
			paths:     []string{"items.0"},
			wantValue: float64(1),
			wantOk:    true,
		},
		{
			name:      "数组访问-多参数",
			jsonData:  `{"data": {"items": ["a", "b", "c"]}}`,
			paths:     []string{"data", "items", "1"},
			wantValue: "b",
			wantOk:    true,
		},
		{
			name:      "返回对象",
			jsonData:  `{"user": {"name": "张三", "age": 20}}`,
			paths:     []string{"user"},
			wantValue: map[string]interface{}{"name": "张三", "age": float64(20)},
			wantOk:    true,
		},
		{
			name:      "返回数组",
			jsonData:  `{"tags": ["go", "sql"]}`,
			paths:     []string{"tags"},
			wantValue: []interface{}{"go", "sql"},
			wantOk:    true,
		},
		{
			name:      "null值",
			jsonData:  `{"value": null}`,
			paths:     []string{"value"},
			wantValue: nil,
			wantOk:    true,
		},
		{
			name:      "空路径",
			jsonData:  `{"name": "张三"}`,
			paths:     []string{},
			wantValue: nil,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data: gjson.Parse(tt.jsonData),
			}

			gotValue, gotOk := getValueByPath(state, tt.paths...)

			if gotOk != tt.wantOk {
				t.Errorf("getValueByPath() ok = %v, want %v", gotOk, tt.wantOk)
				return
			}

			if !compareValues(gotValue, tt.wantValue) {
				t.Errorf("getValueByPath() value = %v (%T), want %v (%T)",
					gotValue, gotValue, tt.wantValue, tt.wantValue)
			}
		})
	}
}

func TestValFunc(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		paths      []string
		wantResult string
		wantArg    interface{}
		wantErr    bool
	}{
		{
			name:       "单层路径-字符串",
			jsonData:   `{"name": "张三"}`,
			paths:      []string{"name"},
			wantResult: "?",
			wantArg:    "张三",
			wantErr:    false,
		},
		{
			name:       "单层路径-数字",
			jsonData:   `{"age": 25}`,
			paths:      []string{"age"},
			wantResult: "?",
			wantArg:    float64(25),
			wantErr:    false,
		},
		{
			name:       "单层路径-布尔值",
			jsonData:   `{"active": true}`,
			paths:      []string{"active"},
			wantResult: "?",
			wantArg:    true,
			wantErr:    false,
		},
		{
			name:       "多层路径-多参数拼接",
			jsonData:   `{"user": {"profile": {"email": "test@example.com"}}}`,
			paths:      []string{"user", "profile", "email"},
			wantResult: "?",
			wantArg:    "test@example.com",
			wantErr:    false,
		},
		{
			name:       "路径不存在-返回nil",
			jsonData:   `{"name": "张三"}`,
			paths:      []string{"age"},
			wantResult: "?",
			wantArg:    nil,
			wantErr:    false,
		},
		{
			name:       "嵌套路径不存在-返回nil",
			jsonData:   `{"user": {"name": "张三"}}`,
			paths:      []string{"user", "email"},
			wantResult: "?",
			wantArg:    nil,
			wantErr:    false,
		},
		{
			name:       "null值",
			jsonData:   `{"value": null}`,
			paths:      []string{"value"},
			wantResult: "?",
			wantArg:    nil,
			wantErr:    false,
		},
		{
			name:       "返回数组",
			jsonData:   `{"tags": ["go", "sql"]}`,
			paths:      []string{"tags"},
			wantResult: "?",
			wantArg:    []interface{}{"go", "sql"},
			wantErr:    false,
		},
		{
			name:       "返回对象",
			jsonData:   `{"user": {"name": "张三", "age": 20}}`,
			paths:      []string{"user"},
			wantResult: "?",
			wantArg:    map[string]interface{}{"name": "张三", "age": float64(20)},
			wantErr:    false,
		},
		{
			name:       "数组索引访问",
			jsonData:   `{"items": [1, 2, 3]}`,
			paths:      []string{"items.1"},
			wantResult: "?",
			wantArg:    float64(2),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data: gjson.Parse(tt.jsonData),
				args: []interface{}{},
			}

			result, err := valFunc(state, tt.paths...)

			if (err != nil) != tt.wantErr {
				t.Errorf("valFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.wantResult {
				t.Errorf("valFunc() result = %v, want %v", result, tt.wantResult)
			}

			if len(state.args) != 1 {
				t.Errorf("valFunc() args length = %v, want 1", len(state.args))
				return
			}

			if !compareValues(state.args[0], tt.wantArg) {
				t.Errorf("valFunc() arg = %v (%T), want %v (%T)",
					state.args[0], state.args[0], tt.wantArg, tt.wantArg)
			}
		})
	}
}

func TestValFuncMultipleCalls(t *testing.T) {
	// 测试多次调用 valFunc 累积参数
	state := &execState{
		data: gjson.Parse(`{"name": "张三", "age": 25, "city": "北京"}`),
		args: []interface{}{},
	}

	// 第一次调用
	result1, err1 := valFunc(state, "name")
	if err1 != nil {
		t.Fatalf("第一次调用 valFunc() error = %v", err1)
	}
	if result1 != "?" {
		t.Errorf("第一次调用 valFunc() result = %v, want ?", result1)
	}

	// 第二次调用
	result2, err2 := valFunc(state, "age")
	if err2 != nil {
		t.Fatalf("第二次调用 valFunc() error = %v", err2)
	}
	if result2 != "?" {
		t.Errorf("第二次调用 valFunc() result = %v, want ?", result2)
	}

	// 第三次调用
	result3, err3 := valFunc(state, "city")
	if err3 != nil {
		t.Fatalf("第三次调用 valFunc() error = %v", err3)
	}
	if result3 != "?" {
		t.Errorf("第三次调用 valFunc() result = %v, want ?", result3)
	}

	// 验证累积的参数
	if len(state.args) != 3 {
		t.Fatalf("valFunc() 累积 args 长度 = %v, want 3", len(state.args))
	}

	expectedArgs := []interface{}{"张三", float64(25), "北京"}
	for i, expected := range expectedArgs {
		if !compareValues(state.args[i], expected) {
			t.Errorf("valFunc() args[%d] = %v (%T), want %v (%T)",
				i, state.args[i], state.args[i], expected, expected)
		}
	}
}

func TestExprFunc(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		paths      []string
		wantResult string
		wantArgs   []interface{}
	}{
		// 参数不足的情况
		{
			name:       "参数不足-0个",
			jsonData:   `{"id": 1}`,
			paths:      []string{},
			wantResult: "",
			wantArgs:   []interface{}{},
		},
		{
			name:       "参数不足-1个",
			jsonData:   `{"id": 1}`,
			paths:      []string{"id"},
			wantResult: "id  ?",
			wantArgs:   []interface{}{nil},
		},
		{
			name:       "参数不足-2个",
			jsonData:   `{"id": 1}`,
			paths:      []string{"id", "="},
			wantResult: "id = ?",
			wantArgs:   []interface{}{nil},
		},

		// 路径不存在的情况
		{
			name:       "路径不存在",
			jsonData:   `{"name": "张三"}`,
			paths:      []string{"id", "=", "age"},
			wantResult: "id = ?",
			wantArgs:   []interface{}{nil},
		},

		// 等于操作符
		{
			name:       "等于操作符-字符串",
			jsonData:   `{"name": "张三"}`,
			paths:      []string{"username", "=", "name"},
			wantResult: "username = ?",
			wantArgs:   []interface{}{"张三"},
		},
		{
			name:       "等于操作符-数字",
			jsonData:   `{"id": 100}`,
			paths:      []string{"user_id", "=", "id"},
			wantResult: "user_id = ?",
			wantArgs:   []interface{}{float64(100)},
		},

		// 比较操作符
		{
			name:       "大于操作符",
			jsonData:   `{"age": 18}`,
			paths:      []string{"user_age", ">", "age"},
			wantResult: "user_age > ?",
			wantArgs:   []interface{}{float64(18)},
		},
		{
			name:       "小于等于操作符",
			jsonData:   `{"price": 99.9}`,
			paths:      []string{"amount", "<=", "price"},
			wantResult: "amount <= ?",
			wantArgs:   []interface{}{float64(99.9)},
		},
		{
			name:       "不等于操作符",
			jsonData:   `{"status": "deleted"}`,
			paths:      []string{"state", "!=", "status"},
			wantResult: "state != ?",
			wantArgs:   []interface{}{"deleted"},
		},
		{
			name:       "LIKE操作符",
			jsonData:   `{"keyword": "%test%"}`,
			paths:      []string{"title", "LIKE", "keyword"},
			wantResult: "title LIKE ?",
			wantArgs:   []interface{}{"%test%"},
		},

		// IN 操作符
		{
			name:       "IN操作符-数组",
			jsonData:   `{"ids": [1, 2, 3]}`,
			paths:      []string{"user_id", "IN", "ids"},
			wantResult: "user_id IN (?, ?, ?)",
			wantArgs:   []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:       "IN操作符-字符串数组",
			jsonData:   `{"names": ["alice", "bob", "charlie"]}`,
			paths:      []string{"username", "IN", "names"},
			wantResult: "username IN (?, ?, ?)",
			wantArgs:   []interface{}{"alice", "bob", "charlie"},
		},
		{
			name:       "IN操作符-单值",
			jsonData:   `{"id": 1}`,
			paths:      []string{"user_id", "IN", "id"},
			wantResult: "user_id IN (?)",
			wantArgs:   []interface{}{float64(1)},
		},
		{
			name:       "in操作符-小写",
			jsonData:   `{"ids": [10, 20]}`,
			paths:      []string{"id", "in", "ids"},
			wantResult: "id in (?, ?)",
			wantArgs:   []interface{}{float64(10), float64(20)},
		},

		// NOT IN 操作符
		{
			name:       "NOT IN操作符",
			jsonData:   `{"excludeIds": [4, 5, 6]}`,
			paths:      []string{"id", "NOT IN", "excludeIds"},
			wantResult: "id NOT IN (?, ?, ?)",
			wantArgs:   []interface{}{float64(4), float64(5), float64(6)},
		},

		// BETWEEN 操作符
		{
			name:       "BETWEEN操作符",
			jsonData:   `{"range": [10, 20]}`,
			paths:      []string{"age", "BETWEEN", "range"},
			wantResult: "age BETWEEN ? AND ?",
			wantArgs:   []interface{}{float64(10), float64(20)},
		},
		{
			name:       "NOT BETWEEN操作符",
			jsonData:   `{"range": [100, 200]}`,
			paths:      []string{"price", "NOT BETWEEN", "range"},
			wantResult: "price NOT BETWEEN ? AND ?",
			wantArgs:   []interface{}{float64(100), float64(200)},
		},

		// 多层路径
		{
			name:       "多层路径访问",
			jsonData:   `{"params": {"filter": {"status": "active"}}}`,
			paths:      []string{"user_status", "=", "params", "filter", "status"},
			wantResult: "user_status = ?",
			wantArgs:   []interface{}{"active"},
		},
		{
			name:       "多层路径-数组",
			jsonData:   `{"query": {"ids": [7, 8, 9]}}`,
			paths:      []string{"record_id", "IN", "query", "ids"},
			wantResult: "record_id IN (?, ?, ?)",
			wantArgs:   []interface{}{float64(7), float64(8), float64(9)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:   gjson.Parse(tt.jsonData),
				args:   []interface{}{},
				errors: []string{},
			}

			result := exprFunc(state, tt.paths...)

			if result != tt.wantResult {
				t.Errorf("exprFunc() result = %q, want %q", result, tt.wantResult)
			}

			if len(state.args) != len(tt.wantArgs) {
				t.Errorf("exprFunc() args length = %v, want %v", len(state.args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(state.args[i], wantArg) {
					t.Errorf("exprFunc() args[%d] = %v (%T), want %v (%T)",
						i, state.args[i], state.args[i], wantArg, wantArg)
				}
			}
		})
	}
}

func TestExprFuncBetweenNotEnoughValues(t *testing.T) {
	// 测试 BETWEEN 操作符值不足的情况
	state := &execState{
		data:   gjson.Parse(`{"value": [10]}`),
		args:   []interface{}{},
		errors: []string{},
	}

	result := exprFunc(state, "age", "BETWEEN", "value")

	if result != "" {
		t.Errorf("exprFunc() BETWEEN with 1 value should return empty, got %q", result)
	}

	if len(state.args) != 0 {
		t.Errorf("exprFunc() BETWEEN with 1 value should not add args, got %v", len(state.args))
	}

	if len(state.errors) != 1 {
		t.Errorf("exprFunc() BETWEEN with 1 value should add error, got %v errors", len(state.errors))
	}
}

func TestExprFuncMultipleCalls(t *testing.T) {
	// 测试多次调用 exprFunc 累积参数
	state := &execState{
		data:   gjson.Parse(`{"name": "张三", "age": 25, "status": ["active", "pending"]}`),
		args:   []interface{}{},
		errors: []string{},
	}

	// 第一次调用 - 等于操作符
	result1 := exprFunc(state, "username", "=", "name")
	if result1 != "username = ?" {
		t.Errorf("第一次调用 exprFunc() result = %q, want %q", result1, "username = ?")
	}

	// 第二次调用 - 大于操作符
	result2 := exprFunc(state, "user_age", ">", "age")
	if result2 != "user_age > ?" {
		t.Errorf("第二次调用 exprFunc() result = %q, want %q", result2, "user_age > ?")
	}

	// 第三次调用 - IN 操作符
	result3 := exprFunc(state, "user_status", "IN", "status")
	if result3 != "user_status IN (?, ?)" {
		t.Errorf("第三次调用 exprFunc() result = %q, want %q", result3, "user_status IN (?, ?)")
	}

	// 验证累积的参数: "张三", 25, "active", "pending"
	expectedArgs := []interface{}{"张三", float64(25), "active", "pending"}
	if len(state.args) != len(expectedArgs) {
		t.Fatalf("exprFunc() 累积 args 长度 = %v, want %v", len(state.args), len(expectedArgs))
	}

	for i, expected := range expectedArgs {
		if !compareValues(state.args[i], expected) {
			t.Errorf("exprFunc() args[%d] = %v (%T), want %v (%T)",
				i, state.args[i], state.args[i], expected, expected)
		}
	}
}

func TestGetValueByPathForTemplate(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		paths     []string
		wantValue interface{}
	}{
		{
			name:      "存在的路径-返回值",
			jsonData:  `{"name": "张三"}`,
			paths:     []string{"name"},
			wantValue: "张三",
		},
		{
			name:      "不存在的路径-返回nil",
			jsonData:  `{"name": "张三"}`,
			paths:     []string{"age"},
			wantValue: nil,
		},
		{
			name:      "多层路径存在",
			jsonData:  `{"user": {"name": "李四"}}`,
			paths:     []string{"user", "name"},
			wantValue: "李四",
		},
		{
			name:      "多层路径不存在",
			jsonData:  `{"user": {"name": "李四"}}`,
			paths:     []string{"user", "age"},
			wantValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data: gjson.Parse(tt.jsonData),
			}

			gotValue := getValueByPathForTemplate(state, tt.paths...)

			if !compareValues(gotValue, tt.wantValue) {
				t.Errorf("getValueByPathForTemplate() = %v (%T), want %v (%T)",
					gotValue, gotValue, tt.wantValue, tt.wantValue)
			}
		})
	}
}

func TestOptionalExprFunc(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		paths      []string
		wantResult string
		wantArgs   []interface{}
		wantErrors int
	}{
		{
			name:       "值存在-生成表达式",
			jsonData:   `{"name": "张三"}`,
			paths:      []string{"username", "=", "name"},
			wantResult: "username = ?",
			wantArgs:   []interface{}{"张三"},
			wantErrors: 0,
		},
		{
			name:       "IN操作符-值存在",
			jsonData:   `{"ids": [1, 2, 3]}`,
			paths:      []string{"user_id", "IN", "ids"},
			wantResult: "user_id IN (?, ?, ?)",
			wantArgs:   []interface{}{float64(1), float64(2), float64(3)},
			wantErrors: 0,
		},
		{
			name:       "BETWEEN操作符-值存在",
			jsonData:   `{"range": [10, 20]}`,
			paths:      []string{"age", "BETWEEN", "range"},
			wantResult: "age BETWEEN ? AND ?",
			wantArgs:   []interface{}{float64(10), float64(20)},
			wantErrors: 0,
		},
		{
			name:       "多层路径-值存在",
			jsonData:   `{"params": {"filter": {"status": "active"}}}`,
			paths:      []string{"user_status", "=", "params", "filter", "status"},
			wantResult: "user_status = ?",
			wantArgs:   []interface{}{"active"},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:   gjson.Parse(tt.jsonData),
				args:   []interface{}{},
				errors: []string{},
			}

			result := optionalExprFunc(state, tt.paths...)

			if result != tt.wantResult {
				t.Errorf("optionalExprFunc() result = %q, want %q", result, tt.wantResult)
			}

			if len(state.args) != len(tt.wantArgs) {
				t.Errorf("optionalExprFunc() args length = %v, want %v", len(state.args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(state.args[i], wantArg) {
					t.Errorf("optionalExprFunc() args[%d] = %v (%T), want %v (%T)",
						i, state.args[i], state.args[i], wantArg, wantArg)
				}
			}

			if len(state.errors) != tt.wantErrors {
				t.Errorf("optionalExprFunc() errors count = %v, want %v", len(state.errors), tt.wantErrors)
			}
		})
	}
}

func TestAndFuncUnit(t *testing.T) {
	tests := []struct {
		name       string
		funcName   string
		conditions []string
		wantResult string
		wantErrors int
	}{
		{
			name:       "两个有效条件",
			funcName:   "and",
			conditions: []string{"a = 1", "b = 2"},
			wantResult: "(a = 1 and b = 2)",
			wantErrors: 0,
		},
		{
			name:       "三个有效条件",
			funcName:   "and",
			conditions: []string{"a = 1", "b = 2", "c = 3"},
			wantResult: "(a = 1 and b = 2 and c = 3)",
			wantErrors: 0,
		},
		{
			name:       "包含空条件-被过滤",
			funcName:   "and",
			conditions: []string{"a = 1", "", "b = 2", "  "},
			wantResult: "(a = 1 and b = 2)",
			wantErrors: 0,
		},
		{
			name:       "单个有效条件",
			funcName:   "and",
			conditions: []string{"a = 1"},
			wantResult: "(a = 1)",
			wantErrors: 0,
		},
		{
			name:       "全部为空条件",
			funcName:   "and",
			conditions: []string{"", "  ", ""},
			wantResult: "",
			wantErrors: 1,
		},
		{
			name:       "无条件",
			funcName:   "and",
			conditions: []string{},
			wantResult: "",
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				errors: []string{},
			}

			result := andFunc(state, tt.funcName, tt.conditions...)

			if result != tt.wantResult {
				t.Errorf("andFunc() result = %q, want %q", result, tt.wantResult)
			}

			if len(state.errors) != tt.wantErrors {
				t.Errorf("andFunc() errors count = %v, want %v", len(state.errors), tt.wantErrors)
			}
		})
	}
}

func TestOrFuncUnit(t *testing.T) {
	tests := []struct {
		name       string
		conditions []string
		wantResult string
		wantErrors int
	}{
		{
			name:       "两个有效条件",
			conditions: []string{"a = 1", "b = 2"},
			wantResult: "(a = 1 or b = 2)",
			wantErrors: 0,
		},
		{
			name:       "三个有效条件",
			conditions: []string{"a = 1", "b = 2", "c = 3"},
			wantResult: "(a = 1 or b = 2 or c = 3)",
			wantErrors: 0,
		},
		{
			name:       "包含空条件-被过滤",
			conditions: []string{"a = 1", "", "b = 2", "  "},
			wantResult: "(a = 1 or b = 2)",
			wantErrors: 0,
		},
		{
			name:       "单个有效条件",
			conditions: []string{"a = 1"},
			wantResult: "(a = 1)",
			wantErrors: 0,
		},
		{
			name:       "全部为空条件",
			conditions: []string{"", "  ", ""},
			wantResult: "",
			wantErrors: 1,
		},
		{
			name:       "无条件",
			conditions: []string{},
			wantResult: "",
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				errors: []string{},
			}

			result := orFunc(state, tt.conditions...)

			if result != tt.wantResult {
				t.Errorf("orFunc() result = %q, want %q", result, tt.wantResult)
			}

			if len(state.errors) != tt.wantErrors {
				t.Errorf("orFunc() errors count = %v, want %v", len(state.errors), tt.wantErrors)
			}
		})
	}
}

func TestValidatorIntFunc(t *testing.T) {
	// 注意: JSON 解析的数字类型是 float64，不是 int
	// 所以 validatorIntFunc 对于 JSON 输入的数字会报错
	// 这个测试用于验证当前的实现行为
	tests := []struct {
		name            string
		jsonData        string
		fieldName       string
		code            string
		msg             string
		paths           []string
		wantResult      string
		wantErrorsCount int
		wantErrorType   string
	}{
		{
			name:            "值不存在-无错误",
			jsonData:        `{"name": "张三"}`,
			fieldName:       "age",
			code:            "AGE_REQUIRED",
			msg:             "年龄必填",
			paths:           []string{"age"},
			wantResult:      "",
			wantErrorsCount: 0,
		},
		{
			name:            "JSON数字被解析为float64-有错误",
			jsonData:        `{"age": 25}`,
			fieldName:       "age",
			code:            "AGE_INVALID",
			msg:             "年龄格式错误",
			paths:           []string{"age"},
			wantResult:      "",
			wantErrorsCount: 1, // JSON 解析数字为 float64，不是 int
			wantErrorType:   ErrValidatorTypeInt,
		},
		{
			name:            "值是浮点数-有错误",
			jsonData:        `{"age": 25.5}`,
			fieldName:       "age",
			code:            "AGE_INVALID",
			msg:             "年龄必须是整数",
			paths:           []string{"age"},
			wantResult:      "",
			wantErrorsCount: 1,
			wantErrorType:   ErrValidatorTypeInt,
		},
		{
			name:            "值是字符串-有错误",
			jsonData:        `{"age": "25"}`,
			fieldName:       "age",
			code:            "AGE_INVALID",
			msg:             "年龄必须是整数",
			paths:           []string{"age"},
			wantResult:      "",
			wantErrorsCount: 1,
			wantErrorType:   ErrValidatorTypeInt,
		},
		{
			name:            "值是布尔值-有错误",
			jsonData:        `{"flag": true}`,
			fieldName:       "flag",
			code:            "FLAG_INVALID",
			msg:             "标志必须是整数",
			paths:           []string{"flag"},
			wantResult:      "",
			wantErrorsCount: 1,
			wantErrorType:   ErrValidatorTypeInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:             gjson.Parse(tt.jsonData),
				validatorsErrors: []*ValidatorError{},
			}

			result := validatorIntFunc(state, tt.fieldName, tt.code, tt.msg, tt.paths...)

			if result != tt.wantResult {
				t.Errorf("validatorIntFunc() result = %q, want %q", result, tt.wantResult)
			}

			if len(state.validatorsErrors) != tt.wantErrorsCount {
				t.Errorf("validatorIntFunc() errors count = %v, want %v", len(state.validatorsErrors), tt.wantErrorsCount)
			}

			if tt.wantErrorsCount > 0 && len(state.validatorsErrors) > 0 {
				if state.validatorsErrors[0].Type != tt.wantErrorType {
					t.Errorf("validatorIntFunc() error type = %q, want %q", state.validatorsErrors[0].Type, tt.wantErrorType)
				}
				if state.validatorsErrors[0].FieldName != tt.fieldName {
					t.Errorf("validatorIntFunc() error fieldName = %q, want %q", state.validatorsErrors[0].FieldName, tt.fieldName)
				}
				if state.validatorsErrors[0].Code != tt.code {
					t.Errorf("validatorIntFunc() error code = %q, want %q", state.validatorsErrors[0].Code, tt.code)
				}
			}
		})
	}
}

func TestValidatorRequiredFunc(t *testing.T) {
	tests := []struct {
		name            string
		jsonData        string
		fieldName       string
		code            string
		msg             string
		paths           []string
		wantErrorsCount int
	}{
		{
			name:            "值存在-无错误",
			jsonData:        `{"name": "张三"}`,
			fieldName:       "name",
			code:            "NAME_REQUIRED",
			msg:             "名称必填",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
		{
			name:            "值不存在-有错误",
			jsonData:        `{"name": "张三"}`,
			fieldName:       "age",
			code:            "AGE_REQUIRED",
			msg:             "年龄必填",
			paths:           []string{"age"},
			wantErrorsCount: 1,
		},
		{
			name:            "嵌套路径存在-无错误",
			jsonData:        `{"user": {"name": "张三"}}`,
			fieldName:       "userName",
			code:            "USER_NAME_REQUIRED",
			msg:             "用户名必填",
			paths:           []string{"user", "name"},
			wantErrorsCount: 0,
		},
		{
			name:            "嵌套路径不存在-有错误",
			jsonData:        `{"user": {"name": "张三"}}`,
			fieldName:       "userAge",
			code:            "USER_AGE_REQUIRED",
			msg:             "用户年龄必填",
			paths:           []string{"user", "age"},
			wantErrorsCount: 1,
		},
		{
			name:            "null值存在-无错误",
			jsonData:        `{"value": null}`,
			fieldName:       "value",
			code:            "VALUE_REQUIRED",
			msg:             "值必填",
			paths:           []string{"value"},
			wantErrorsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:             gjson.Parse(tt.jsonData),
				validatorsErrors: []*ValidatorError{},
			}

			result := validatorRequiredFunc(state, tt.fieldName, tt.code, tt.msg, tt.paths...)

			if result != "" {
				t.Errorf("validatorRequiredFunc() result = %q, want empty string", result)
			}

			if len(state.validatorsErrors) != tt.wantErrorsCount {
				t.Errorf("validatorRequiredFunc() errors count = %v, want %v", len(state.validatorsErrors), tt.wantErrorsCount)
			}

			if tt.wantErrorsCount > 0 && len(state.validatorsErrors) > 0 {
				if state.validatorsErrors[0].Type != ErrValidatorRequired {
					t.Errorf("validatorRequiredFunc() error type = %q, want %q", state.validatorsErrors[0].Type, ErrValidatorRequired)
				}
			}
		})
	}
}

func TestValidatorStrFunc(t *testing.T) {
	tests := []struct {
		name            string
		jsonData        string
		fieldName       string
		code            string
		msg             string
		paths           []string
		wantErrorsCount int
	}{
		{
			name:            "值不存在-无错误",
			jsonData:        `{"age": 25}`,
			fieldName:       "name",
			code:            "NAME_INVALID",
			msg:             "名称格式错误",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
		{
			name:            "值是字符串-无错误",
			jsonData:        `{"name": "张三"}`,
			fieldName:       "name",
			code:            "NAME_INVALID",
			msg:             "名称格式错误",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
		{
			name:            "值是数字-有错误",
			jsonData:        `{"name": 123}`,
			fieldName:       "name",
			code:            "NAME_INVALID",
			msg:             "名称必须是字符串",
			paths:           []string{"name"},
			wantErrorsCount: 1,
		},
		{
			name:            "值是布尔值-有错误",
			jsonData:        `{"active": true}`,
			fieldName:       "active",
			code:            "ACTIVE_INVALID",
			msg:             "状态必须是字符串",
			paths:           []string{"active"},
			wantErrorsCount: 1,
		},
		{
			name:            "空字符串-无错误",
			jsonData:        `{"name": ""}`,
			fieldName:       "name",
			code:            "NAME_INVALID",
			msg:             "名称格式错误",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:             gjson.Parse(tt.jsonData),
				validatorsErrors: []*ValidatorError{},
			}

			result := validatorStrFunc(state, tt.fieldName, tt.code, tt.msg, tt.paths...)

			if result != "" {
				t.Errorf("validatorStrFunc() result = %q, want empty string", result)
			}

			if len(state.validatorsErrors) != tt.wantErrorsCount {
				t.Errorf("validatorStrFunc() errors count = %v, want %v", len(state.validatorsErrors), tt.wantErrorsCount)
			}

			if tt.wantErrorsCount > 0 && len(state.validatorsErrors) > 0 {
				if state.validatorsErrors[0].Type != ErrValidatorTypeStr {
					t.Errorf("validatorStrFunc() error type = %q, want %q", state.validatorsErrors[0].Type, ErrValidatorTypeStr)
				}
			}
		})
	}
}

func TestValidatorStrLenFunc(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name            string
		jsonData        string
		min             *int
		max             *int
		fieldName       string
		code            string
		msg             string
		paths           []string
		wantErrorsCount int
	}{
		{
			name:            "值不存在-无错误",
			jsonData:        `{"age": 25}`,
			min:             intPtr(1),
			max:             intPtr(10),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度错误",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
		{
			name:            "长度在范围内-无错误",
			jsonData:        `{"name": "张三"}`,
			min:             intPtr(1),
			max:             intPtr(10),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度错误",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
		{
			name:            "长度小于最小值-有错误",
			jsonData:        `{"name": "a"}`,
			min:             intPtr(5),
			max:             intPtr(10),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度必须在5-10之间",
			paths:           []string{"name"},
			wantErrorsCount: 1,
		},
		{
			name:            "长度大于最大值-有错误",
			jsonData:        `{"name": "这是一个非常长的名称"}`,
			min:             intPtr(1),
			max:             intPtr(5),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度必须在1-5之间",
			paths:           []string{"name"},
			wantErrorsCount: 1,
		},
		{
			name:            "值不是字符串-有错误",
			jsonData:        `{"name": 123}`,
			min:             intPtr(1),
			max:             intPtr(10),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称必须是字符串",
			paths:           []string{"name"},
			wantErrorsCount: 1,
		},
		{
			name:            "只设置最小值-通过",
			jsonData:        `{"name": "abcdef"}`,
			min:             intPtr(3),
			max:             nil,
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度不足",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
		{
			name:            "只设置最小值-失败",
			jsonData:        `{"name": "ab"}`,
			min:             intPtr(3),
			max:             nil,
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度不足",
			paths:           []string{"name"},
			wantErrorsCount: 1,
		},
		{
			name:            "只设置最大值-通过",
			jsonData:        `{"name": "abc"}`,
			min:             nil,
			max:             intPtr(5),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度过长",
			paths:           []string{"name"},
			wantErrorsCount: 0,
		},
		{
			name:            "只设置最大值-失败",
			jsonData:        `{"name": "abcdefgh"}`,
			min:             nil,
			max:             intPtr(5),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称长度过长",
			paths:           []string{"name"},
			wantErrorsCount: 1,
		},
		{
			name:            "空字符串-长度为0",
			jsonData:        `{"name": ""}`,
			min:             intPtr(1),
			max:             intPtr(10),
			fieldName:       "name",
			code:            "NAME_LEN",
			msg:             "名称不能为空",
			paths:           []string{"name"},
			wantErrorsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:             gjson.Parse(tt.jsonData),
				validatorsErrors: []*ValidatorError{},
			}

			result := validatorStrLenFunc(state, tt.min, tt.max, tt.fieldName, tt.code, tt.msg, tt.paths...)

			if result != "" {
				t.Errorf("validatorStrLenFunc() result = %q, want empty string", result)
			}

			if len(state.validatorsErrors) != tt.wantErrorsCount {
				t.Errorf("validatorStrLenFunc() errors count = %v, want %v", len(state.validatorsErrors), tt.wantErrorsCount)
			}

			if tt.wantErrorsCount > 0 && len(state.validatorsErrors) > 0 {
				if state.validatorsErrors[0].Type != ErrValidatorTypeStrLen {
					t.Errorf("validatorStrLenFunc() error type = %q, want %q", state.validatorsErrors[0].Type, ErrValidatorTypeStrLen)
				}
			}
		})
	}
}

func TestValidatorArrLenFunc(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name            string
		jsonData        string
		min             *int
		max             *int
		fieldName       string
		code            string
		msg             string
		paths           []string
		wantErrorsCount int
	}{
		{
			name:            "值不存在-无错误",
			jsonData:        `{"name": "张三"}`,
			min:             intPtr(1),
			max:             intPtr(10),
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "数组长度错误",
			paths:           []string{"items"},
			wantErrorsCount: 0,
		},
		{
			name:            "长度在范围内-无错误",
			jsonData:        `{"items": [1, 2, 3]}`,
			min:             intPtr(1),
			max:             intPtr(5),
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "数组长度错误",
			paths:           []string{"items"},
			wantErrorsCount: 0,
		},
		{
			name:            "长度小于最小值-有错误",
			jsonData:        `{"items": [1]}`,
			min:             intPtr(3),
			max:             intPtr(10),
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "数组长度必须在3-10之间",
			paths:           []string{"items"},
			wantErrorsCount: 1,
		},
		{
			name:            "长度大于最大值-有错误",
			jsonData:        `{"items": [1, 2, 3, 4, 5, 6]}`,
			min:             intPtr(1),
			max:             intPtr(3),
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "数组长度必须在1-3之间",
			paths:           []string{"items"},
			wantErrorsCount: 1,
		},
		{
			name:            "值不是数组-有错误",
			jsonData:        `{"items": "not array"}`,
			min:             intPtr(1),
			max:             intPtr(10),
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "必须是数组",
			paths:           []string{"items"},
			wantErrorsCount: 1,
		},
		{
			name:            "空数组-长度为0",
			jsonData:        `{"items": []}`,
			min:             intPtr(1),
			max:             intPtr(10),
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "数组不能为空",
			paths:           []string{"items"},
			wantErrorsCount: 1,
		},
		{
			name:            "只设置最小值-通过",
			jsonData:        `{"items": [1, 2, 3, 4]}`,
			min:             intPtr(2),
			max:             nil,
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "数组长度不足",
			paths:           []string{"items"},
			wantErrorsCount: 0,
		},
		{
			name:            "只设置最大值-通过",
			jsonData:        `{"items": [1, 2]}`,
			min:             nil,
			max:             intPtr(5),
			fieldName:       "items",
			code:            "ITEMS_LEN",
			msg:             "数组长度过长",
			paths:           []string{"items"},
			wantErrorsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:             gjson.Parse(tt.jsonData),
				validatorsErrors: []*ValidatorError{},
			}

			result := validatorArrLenFunc(state, tt.min, tt.max, tt.fieldName, tt.code, tt.msg, tt.paths...)

			if result != "" {
				t.Errorf("validatorArrLenFunc() result = %q, want empty string", result)
			}

			if len(state.validatorsErrors) != tt.wantErrorsCount {
				t.Errorf("validatorArrLenFunc() errors count = %v, want %v", len(state.validatorsErrors), tt.wantErrorsCount)
			}

			if tt.wantErrorsCount > 0 && len(state.validatorsErrors) > 0 {
				if state.validatorsErrors[0].Type != ErrValidatorTypeArrLen {
					t.Errorf("validatorArrLenFunc() error type = %q, want %q", state.validatorsErrors[0].Type, ErrValidatorTypeArrLen)
				}
			}
		})
	}
}

func TestValidatorRegFunc(t *testing.T) {
	tests := []struct {
		name            string
		jsonData        string
		pattern         string
		fieldName       string
		code            string
		msg             string
		paths           []string
		wantErrorsCount int
	}{
		{
			name:            "值不存在-无错误",
			jsonData:        `{"name": "张三"}`,
			pattern:         `^[a-z]+$`,
			fieldName:       "email",
			code:            "EMAIL_INVALID",
			msg:             "邮箱格式错误",
			paths:           []string{"email"},
			wantErrorsCount: 0,
		},
		{
			name:            "匹配正则-无错误",
			jsonData:        `{"email": "test@example.com"}`,
			pattern:         `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			fieldName:       "email",
			code:            "EMAIL_INVALID",
			msg:             "邮箱格式错误",
			paths:           []string{"email"},
			wantErrorsCount: 0,
		},
		{
			name:            "不匹配正则-有错误",
			jsonData:        `{"email": "invalid-email"}`,
			pattern:         `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			fieldName:       "email",
			code:            "EMAIL_INVALID",
			msg:             "邮箱格式错误",
			paths:           []string{"email"},
			wantErrorsCount: 1,
		},
		{
			name:            "值不是字符串-有错误",
			jsonData:        `{"phone": 12345678}`,
			pattern:         `^\d{11}$`,
			fieldName:       "phone",
			code:            "PHONE_INVALID",
			msg:             "手机号格式错误",
			paths:           []string{"phone"},
			wantErrorsCount: 1,
		},
		{
			name:            "手机号正则匹配",
			jsonData:        `{"phone": "13812345678"}`,
			pattern:         `^1[3-9]\d{9}$`,
			fieldName:       "phone",
			code:            "PHONE_INVALID",
			msg:             "手机号格式错误",
			paths:           []string{"phone"},
			wantErrorsCount: 0,
		},
		{
			name:            "手机号正则不匹配",
			jsonData:        `{"phone": "12345678901"}`,
			pattern:         `^1[3-9]\d{9}$`,
			fieldName:       "phone",
			code:            "PHONE_INVALID",
			msg:             "手机号格式错误",
			paths:           []string{"phone"},
			wantErrorsCount: 1,
		},
		{
			name:            "空字符串-不匹配",
			jsonData:        `{"code": ""}`,
			pattern:         `^[A-Z]{3}$`,
			fieldName:       "code",
			code:            "CODE_INVALID",
			msg:             "编码格式错误",
			paths:           []string{"code"},
			wantErrorsCount: 1,
		},
		{
			name:            "嵌套路径正则验证",
			jsonData:        `{"user": {"code": "ABC"}}`,
			pattern:         `^[A-Z]{3}$`,
			fieldName:       "userCode",
			code:            "CODE_INVALID",
			msg:             "编码格式错误",
			paths:           []string{"user", "code"},
			wantErrorsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:             gjson.Parse(tt.jsonData),
				validatorsErrors: []*ValidatorError{},
			}

			result := validatorRegFunc(state, tt.pattern, tt.fieldName, tt.code, tt.msg, tt.paths...)

			if result != "" {
				t.Errorf("validatorRegFunc() result = %q, want empty string", result)
			}

			if len(state.validatorsErrors) != tt.wantErrorsCount {
				t.Errorf("validatorRegFunc() errors count = %v, want %v", len(state.validatorsErrors), tt.wantErrorsCount)
			}

			if tt.wantErrorsCount > 0 && len(state.validatorsErrors) > 0 {
				if state.validatorsErrors[0].Type != ErrValidatorTypeReg {
					t.Errorf("validatorRegFunc() error type = %q, want %q", state.validatorsErrors[0].Type, ErrValidatorTypeReg)
				}
			}
		})
	}
}

func TestBuildExprWithDifferentArrayTypes(t *testing.T) {
	// 测试 buildExpr 对不同类型数组的处理
	tests := []struct {
		name       string
		val        interface{}
		wantResult string
		wantArgs   []interface{}
	}{
		{
			name:       "[]string类型",
			val:        []string{"a", "b", "c"},
			wantResult: "field IN (?, ?, ?)",
			wantArgs:   []interface{}{"a", "b", "c"},
		},
		{
			name:       "[]int类型",
			val:        []int{1, 2, 3},
			wantResult: "field IN (?, ?, ?)",
			wantArgs:   []interface{}{1, 2, 3},
		},
		{
			name:       "[]int64类型",
			val:        []int64{100, 200, 300},
			wantResult: "field IN (?, ?, ?)",
			wantArgs:   []interface{}{int64(100), int64(200), int64(300)},
		},
		{
			name:       "[]interface{}类型",
			val:        []interface{}{"x", 1, true},
			wantResult: "field IN (?, ?, ?)",
			wantArgs:   []interface{}{"x", 1, true},
		},
		{
			name:       "单值转数组",
			val:        "single",
			wantResult: "field IN (?)",
			wantArgs:   []interface{}{"single"},
		},
		{
			name:       "nil值-required",
			val:        nil,
			wantResult: "field IN (?)",
			wantArgs:   []interface{}{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				args:   []interface{}{},
				errors: []string{},
			}

			result := buildExpr(state, "field", "IN", true, tt.val)

			if result != tt.wantResult {
				t.Errorf("buildExpr() result = %q, want %q", result, tt.wantResult)
			}

			if len(state.args) != len(tt.wantArgs) {
				t.Errorf("buildExpr() args length = %v, want %v", len(state.args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(state.args[i], wantArg) {
					t.Errorf("buildExpr() args[%d] = %v (%T), want %v (%T)",
						i, state.args[i], state.args[i], wantArg, wantArg)
				}
			}
		})
	}
}

func TestExecStateAddError(t *testing.T) {
	state := &execState{
		errors: []string{},
	}

	state.addError("error 1")
	state.addError("error 2")

	if len(state.errors) != 2 {
		t.Errorf("addError() errors count = %v, want 2", len(state.errors))
	}

	if state.errors[0] != "error 1" || state.errors[1] != "error 2" {
		t.Errorf("addError() errors = %v, want [error 1, error 2]", state.errors)
	}
}

func TestExecStateAddValidatorError(t *testing.T) {
	state := &execState{
		validatorsErrors: []*ValidatorError{},
	}

	err1 := NewValidatorError(ErrValidatorRequired, "name", "NAME_REQUIRED", "名称必填")
	err2 := NewValidatorError(ErrValidatorTypeStr, "age", "AGE_INVALID", "年龄格式错误")

	state.addValidatorError(err1)
	state.addValidatorError(err2)

	if len(state.validatorsErrors) != 2 {
		t.Errorf("addValidatorError() errors count = %v, want 2", len(state.validatorsErrors))
	}

	if state.validatorsErrors[0].FieldName != "name" || state.validatorsErrors[1].FieldName != "age" {
		t.Errorf("addValidatorError() not working correctly")
	}
}

// compareValues 比较两个值是否相等（处理 map 和 slice 的深度比较）
func compareValues(got, want interface{}) bool {
	if got == nil && want == nil {
		return true
	}
	if got == nil || want == nil {
		return false
	}

	switch w := want.(type) {
	case map[string]interface{}:
		g, ok := got.(map[string]interface{})
		if !ok {
			return false
		}
		if len(g) != len(w) {
			return false
		}
		for k, v := range w {
			if gv, exists := g[k]; !exists || !compareValues(gv, v) {
				return false
			}
		}
		return true
	case []interface{}:
		g, ok := got.([]interface{})
		if !ok {
			return false
		}
		if len(g) != len(w) {
			return false
		}
		for i, v := range w {
			if !compareValues(g[i], v) {
				return false
			}
		}
		return true
	default:
		return got == want
	}
}
