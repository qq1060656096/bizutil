package qsql

import (
	"testing"

	"github.com/tidwall/gjson"
)

// TestAndFunc 测试 andFunc 函数
func TestAndFunc(t *testing.T) {
	tests := []struct {
		name       string
		conditions []string
		wantResult string
		wantError  bool
	}{
		{
			name:       "无条件",
			conditions: []string{},
			wantResult: "",
			wantError:  true,
		},
		{
			name:       "全部空条件",
			conditions: []string{"", "  ", ""},
			wantResult: "",
			wantError:  true,
		},
		{
			name:       "单个条件",
			conditions: []string{"name = ?"},
			wantResult: "(name = ?)",
			wantError:  false,
		},
		{
			name:       "两个条件",
			conditions: []string{"name = ?", "age > ?"},
			wantResult: "(name = ? and age > ?)",
			wantError:  false,
		},
		{
			name:       "三个条件",
			conditions: []string{"a = ?", "b = ?", "c = ?"},
			wantResult: "(a = ? and b = ? and c = ?)",
			wantError:  false,
		},
		{
			name:       "混合空条件",
			conditions: []string{"name = ?", "", "age > ?", "  "},
			wantResult: "(name = ? and age > ?)",
			wantError:  false,
		},
		{
			name:       "只有一个有效条件",
			conditions: []string{"", "name = ?", ""},
			wantResult: "(name = ?)",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:   gjson.Parse(`{}`),
				args:   []interface{}{},
				errors: []string{},
			}

			// andFunc 需要一个 funcName 作为第一个参数
			result := andFunc(state, "and", tt.conditions...)

			if result != tt.wantResult {
				t.Errorf("andFunc() = %q, want %q", result, tt.wantResult)
			}

			hasError := len(state.errors) > 0
			if hasError != tt.wantError {
				t.Errorf("andFunc() hasError = %v, want %v, errors: %v", hasError, tt.wantError, state.errors)
			}
		})
	}
}

// TestOrFunc 测试 orFunc 函数
func TestOrFunc(t *testing.T) {
	tests := []struct {
		name       string
		conditions []string
		wantResult string
		wantError  bool
	}{
		{
			name:       "无条件",
			conditions: []string{},
			wantResult: "",
			wantError:  true,
		},
		{
			name:       "全部空条件",
			conditions: []string{"", "  ", ""},
			wantResult: "",
			wantError:  true,
		},
		{
			name:       "单个条件",
			conditions: []string{"name = ?"},
			wantResult: "(name = ?)",
			wantError:  false,
		},
		{
			name:       "两个条件",
			conditions: []string{"name = ?", "age > ?"},
			wantResult: "(name = ? or age > ?)",
			wantError:  false,
		},
		{
			name:       "三个条件",
			conditions: []string{"a = ?", "b = ?", "c = ?"},
			wantResult: "(a = ? or b = ? or c = ?)",
			wantError:  false,
		},
		{
			name:       "混合空条件",
			conditions: []string{"name = ?", "", "age > ?", "  "},
			wantResult: "(name = ? or age > ?)",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &execState{
				data:   gjson.Parse(`{}`),
				args:   []interface{}{},
				errors: []string{},
			}

			result := orFunc(state, tt.conditions...)

			if result != tt.wantResult {
				t.Errorf("orFunc() = %q, want %q", result, tt.wantResult)
			}

			hasError := len(state.errors) > 0
			if hasError != tt.wantError {
				t.Errorf("orFunc() hasError = %v, want %v, errors: %v", hasError, tt.wantError, state.errors)
			}
		})
	}
}

// TestEngineExecuteExpr 测试通过 Engine 执行 expr 表达式
func TestEngineExecuteExpr(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		{
			name:       "expr-等于操作符",
			template:   `SELECT * FROM users WHERE {expr . "name" "=" "params.name"}`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `SELECT * FROM users WHERE name = ?`,
			wantArgs:   []interface{}{"张三"},
		},
		{
			name:       "expr-大于操作符",
			template:   `SELECT * FROM users WHERE {expr . "age" ">" "params.age"}`,
			paramsJSON: `{"params": {"age": 18}}`,
			wantSQL:    `SELECT * FROM users WHERE age > ?`,
			wantArgs:   []interface{}{float64(18)},
		},
		{
			name:       "expr-LIKE操作符",
			template:   `SELECT * FROM users WHERE {expr . "name" "LIKE" "params.keyword"}`,
			paramsJSON: `{"params": {"keyword": "%test%"}}`,
			wantSQL:    `SELECT * FROM users WHERE name LIKE ?`,
			wantArgs:   []interface{}{"%test%"},
		},
		{
			name:       "expr-IN操作符",
			template:   `SELECT * FROM users WHERE {expr . "id" "IN" "params.ids"}`,
			paramsJSON: `{"params": {"ids": [1, 2, 3]}}`,
			wantSQL:    `SELECT * FROM users WHERE id IN (?, ?, ?)`,
			wantArgs:   []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:       "expr-NOT IN操作符",
			template:   `SELECT * FROM users WHERE {expr . "id" "NOT IN" "params.excludeIds"}`,
			paramsJSON: `{"params": {"excludeIds": [4, 5]}}`,
			wantSQL:    `SELECT * FROM users WHERE id NOT IN (?, ?)`,
			wantArgs:   []interface{}{float64(4), float64(5)},
		},
		{
			name:       "expr-BETWEEN操作符",
			template:   `SELECT * FROM users WHERE {expr . "age" "BETWEEN" "params.ageRange"}`,
			paramsJSON: `{"params": {"ageRange": [18, 30]}}`,
			wantSQL:    `SELECT * FROM users WHERE age BETWEEN ? AND ?`,
			wantArgs:   []interface{}{float64(18), float64(30)},
		},
		{
			name:       "expr-路径不存在返回空",
			template:   `SELECT * FROM users WHERE 1=1 {expr . "name" "=" "params.notExist"}`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1 name = ?`,
			wantArgs:   []interface{}{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestEngineExecuteVal 测试通过 Engine 执行 val 函数
func TestEngineExecuteVal(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		{
			name:       "val-字符串值",
			template:   `INSERT INTO users (name) VALUES ({val . "params.name"})`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `INSERT INTO users (name) VALUES (?)`,
			wantArgs:   []interface{}{"张三"},
		},
		{
			name:       "val-数字值",
			template:   `INSERT INTO users (age) VALUES ({val . "params.age"})`,
			paramsJSON: `{"params": {"age": 25}}`,
			wantSQL:    `INSERT INTO users (age) VALUES (?)`,
			wantArgs:   []interface{}{float64(25)},
		},
		{
			name:       "val-布尔值",
			template:   `UPDATE users SET active = {val . "params.active"}`,
			paramsJSON: `{"params": {"active": true}}`,
			wantSQL:    `UPDATE users SET active = ?`,
			wantArgs:   []interface{}{true},
		},
		{
			name:       "val-多个值",
			template:   `INSERT INTO users (name, age) VALUES ({val . "params.name"}, {val . "params.age"})`,
			paramsJSON: `{"params": {"name": "李四", "age": 30}}`,
			wantSQL:    `INSERT INTO users (name, age) VALUES (?, ?)`,
			wantArgs:   []interface{}{"李四", float64(30)},
		},
		{
			name:       "val-路径不存在返回nil",
			template:   `INSERT INTO users (name) VALUES ({val . "params.notExist"})`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `INSERT INTO users (name) VALUES (?)`,
			wantArgs:   []interface{}{nil},
		},
		{
			name:       "val-嵌套路径",
			template:   `SELECT * FROM users WHERE id = {val . "params.user.id"}`,
			paramsJSON: `{"params": {"user": {"id": 100}}}`,
			wantSQL:    `SELECT * FROM users WHERE id = ?`,
			wantArgs:   []interface{}{float64(100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestEngineExecuteAnd 测试通过 Engine 执行 and 逻辑组合
func TestEngineExecuteAnd(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		{
			name: "and-两个expr条件",
			template: `SELECT * FROM users WHERE {and . "and"
				(expr . "name" "=" "params.name")
				(expr . "age" ">" "params.age")
			}`,
			paramsJSON: `{"params": {"name": "张三", "age": 18}}`,
			wantSQL:    `SELECT * FROM users WHERE (name = ? and age > ?)`,
			wantArgs:   []interface{}{"张三", float64(18)},
		},
		{
			name: "and-三个expr条件",
			template: `SELECT * FROM users WHERE {and . "and"
				(expr . "name" "=" "params.name")
				(expr . "age" ">" "params.age")
				(expr . "status" "=" "params.status")
			}`,
			paramsJSON: `{"params": {"name": "张三", "age": 18, "status": "active"}}`,
			wantSQL:    `SELECT * FROM users WHERE (name = ? and age > ? and status = ?)`,
			wantArgs:   []interface{}{"张三", float64(18), "active"},
		},
		{
			name: "and-部分条件为空",
			template: `SELECT * FROM users WHERE {and . "and"
				(expr . "name" "=" "params.name")
				(expr . "age" ">" "params.age")
			}`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `SELECT * FROM users WHERE (name = ? and age > ?)`,
			wantArgs:   []interface{}{"张三", nil},
		},
		{
			name: "and-所有条件为空返回null",
			template: `SELECT * FROM users WHERE {and . "and"
				(expr . "name" "=" "params.name")
				(expr . "age" ">" "params.age")
			}`,
			paramsJSON: `{"params": {}}`,
			wantSQL:    `SELECT * FROM users WHERE (name = ? and age > ?)`,
			wantArgs:   []interface{}{nil, nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestEngineExecuteOr 测试通过 Engine 执行 or 逻辑组合
func TestEngineExecuteOr(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		{
			name: "or-两个expr条件",
			template: `SELECT * FROM users WHERE {or .
				(expr . "name" "=" "params.name")
				(expr . "email" "=" "params.email")
			}`,
			paramsJSON: `{"params": {"name": "张三", "email": "test@example.com"}}`,
			wantSQL:    `SELECT * FROM users WHERE (name = ? or email = ?)`,
			wantArgs:   []interface{}{"张三", "test@example.com"},
		},
		{
			name: "or-三个expr条件",
			template: `SELECT * FROM users WHERE {or .
				(expr . "name" "LIKE" "params.keyword")
				(expr . "email" "LIKE" "params.keyword")
				(expr . "phone" "LIKE" "params.keyword")
			}`,
			paramsJSON: `{"params": {"keyword": "%test%"}}`,
			wantSQL:    `SELECT * FROM users WHERE (name LIKE ? or email LIKE ? or phone LIKE ?)`,
			wantArgs:   []interface{}{"%test%", "%test%", "%test%"},
		},
		{
			name: "or-部分条件为空",
			template: `SELECT * FROM users WHERE {or .
				(expr . "name" "=" "params.name")
				(expr . "email" "=" "params.email")
			}`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `SELECT * FROM users WHERE (name = ? or email = ?)`,
			wantArgs:   []interface{}{"张三", nil},
		},
		{
			name: "or-所有条件为空返回1=1",
			template: `SELECT * FROM users WHERE {or .
				(expr . "name" "=" "params.name")
			}`,
			paramsJSON: `{"params": {}}`,
			wantSQL:    `SELECT * FROM users WHERE (name = ?)`,
			wantArgs:   []interface{}{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestEngineExecuteRange 测试通过 Engine 执行 range 迭代
func TestEngineExecuteRange(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		{
			name:       "range-遍历数组生成IN条件",
			template:   `SELECT * FROM users WHERE {expr . "id" "IN" "params.ids"}`,
			paramsJSON: `{"params": {"ids": [1, 2, 3]}}`,
			wantSQL:    `SELECT * FROM users WHERE id IN (?, ?, ?)`,
			wantArgs:   []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:       "range-空数组",
			template:   `SELECT * FROM users WHERE 1=1{range $item := (getValue . "params.items")} AND name = ?{end}`,
			paramsJSON: `{"params": {"items": []}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1`,
			wantArgs:   []interface{}{},
		},
		{
			name:       "range-遍历数组生成多个值",
			template:   `SELECT * FROM users WHERE id IN ({range $i, $v := (getValue . "params.ids")}{if $i}, {end}?{end})`,
			paramsJSON: `{"params": {"ids": [1, 2, 3]}}`,
			wantSQL:    `SELECT * FROM users WHERE id IN (?, ?, ?)`,
			wantArgs:   []interface{}{},
		},
		{
			name:       "range+expr-单个遍历生成条件",
			template:   `SELECT * FROM users WHERE 1=1{range $i, $_ := (getValue $ "params.filters")} AND {expr $ "name" "=" (printf "params.filters.%d.value" $i)}{end}`,
			paramsJSON: `{"params": {"filters": [{"field": "name", "value": "张三"}]}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1 AND name = ?`,
			wantArgs:   []interface{}{"张三"},
		},
		{
			name:       "range+expr-多个遍历生成条件",
			template:   `SELECT * FROM users WHERE 1=1{range $i, $_ := (getValue $ "params.filters")} AND {expr $ "status" "=" (printf "params.filters.%d.status" $i)}{end}`,
			paramsJSON: `{"params": {"filters": [{"status": "active"}, {"status": "pending"}, {"status": "completed"}]}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1 AND status = ? AND status = ? AND status = ?`,
			wantArgs:   []interface{}{"active", "pending", "completed"},
		},
		{
			name:       "range+expr-遍历不同操作符",
			template:   `SELECT * FROM orders WHERE 1=1{range $i, $_ := (getValue $ "params.conditions")} AND {expr $ "amount" ">" (printf "params.conditions.%d.min" $i)}{end}`,
			paramsJSON: `{"params": {"conditions": [{"min": 100}, {"min": 200}]}}`,
			wantSQL:    `SELECT * FROM orders WHERE 1=1 AND amount > ? AND amount > ?`,
			wantArgs:   []interface{}{float64(100), float64(200)},
		},
		{
			name:       "range+expr-遍历空数组",
			template:   `SELECT * FROM users WHERE 1=1{range $i, $_ := (getValue $ "params.filters")} AND {expr $ "name" "=" (printf "params.filters.%d.value" $i)}{end}`,
			paramsJSON: `{"params": {"filters": []}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1`,
			wantArgs:   []interface{}{},
		},
		{
			name:       "range+expr-多个range块各含expr",
			template:   `SELECT * FROM users WHERE 1=1{range $i, $_ := (getValue $ "params.users")} AND {expr $ "user_id" "=" (printf "params.users.%d.id" $i)}{end}{range $j, $_ := (getValue $ "params.roles")} AND {expr $ "role" "=" (printf "params.roles.%d.name" $j)}{end}`,
			paramsJSON: `{"params": {"users": [{"id": 1}, {"id": 2}], "roles": [{"name": "admin"}]}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1 AND user_id = ? AND user_id = ? AND role = ?`,
			wantArgs:   []interface{}{float64(1), float64(2), "admin"},
		},
		{
			name: "range+val+expr-订单商品复杂查询",
			template: `select * from business_orders a
left join business_orders_list b on a.company_id = b.company_id and a.orders_id = b.orders_id
where a.company_id = {val $ "params.company_id"} and (
{range $i, $_ := (getValue $ "params.goods")}{if $i} and {end}(b.goods_id = {val $ (printf "params.goods.%d.goods_id" $i)} and b.options_id = {val $ (printf "params.goods.%d.options_id" $i)}){end}
) and (
{range $i, $_ := (getValue $ "params.orders_nums")}{if $i} or {end}{expr $ "a.orders_num" "like" (printf "params.orders_nums.%d" $i)}{end}
)
group by b.orders_list_id
limit 0,10`,
			paramsJSON: `{"params": {"company_id": 1001, "goods": [{"goods_id": 101, "options_id": 201}, {"goods_id": 102, "options_id": 202}], "orders_nums": ["%ORD001%", "%ORD002%"]}}`,
			wantSQL:    `select * from business_orders a left join business_orders_list b on a.company_id = b.company_id and a.orders_id = b.orders_id where a.company_id = ? and ( (b.goods_id = ? and b.options_id = ?) and (b.goods_id = ? and b.options_id = ?) ) and ( a.orders_num like ? or a.orders_num like ? ) group by b.orders_list_id limit 0,10`,
			wantArgs:   []interface{}{float64(1001), float64(101), float64(201), float64(102), float64(202), "%ORD001%", "%ORD002%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestEngineExecuteNestedAndOr 测试嵌套的 and/or 逻辑组合
func TestEngineExecuteNestedAndOr(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		{
			name: "嵌套-and包含or",
			template: `SELECT * FROM users WHERE {and . "and"
				(expr . "status" "=" "params.status")
				(or .
					(expr . "name" "LIKE" "params.keyword")
					(expr . "email" "LIKE" "params.keyword")
				)
			}`,
			paramsJSON: `{"params": {"status": "active", "keyword": "%test%"}}`,
			wantSQL:    `SELECT * FROM users WHERE (status = ? and (name LIKE ? or email LIKE ?))`,
			wantArgs:   []interface{}{"active", "%test%", "%test%"},
		},
		{
			name: "嵌套-or包含and",
			template: `SELECT * FROM users WHERE {or .
				(and . "and"
					(expr . "type" "=" "params.type1")
					(expr . "level" ">" "params.level1")
				)
				(and . "and"
					(expr . "type" "=" "params.type2")
					(expr . "level" ">" "params.level2")
				)
			}`,
			paramsJSON: `{"params": {"type1": "vip", "level1": 5, "type2": "admin", "level2": 3}}`,
			wantSQL:    `SELECT * FROM users WHERE ((type = ? and level > ?) or (type = ? and level > ?))`,
			wantArgs:   []interface{}{"vip", float64(5), "admin", float64(3)},
		},
		{
			name: "嵌套-部分条件缺失",
			template: `SELECT * FROM users WHERE {and . "and"
				(expr . "status" "=" "params.status")
				(or .
					(expr . "name" "LIKE" "params.keyword")
					(expr . "email" "LIKE" "params.keyword")
				)
			}`,
			paramsJSON: `{"params": {"status": "active"}}`,
			wantSQL:    `SELECT * FROM users WHERE (status = ? and (name LIKE ? or email LIKE ?))`,
			wantArgs:   []interface{}{"active", nil, nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestEngineExecuteInvalidJSON 测试无效的 JSON 输入
func TestEngineExecuteInvalidJSON(t *testing.T) {
	engine := NewEngine()
	if err := engine.Parse("test", `SELECT * FROM users WHERE id = {val "params.id" .}`); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, err := engine.Execute(`{invalid json}`)
	if err == nil {
		t.Error("Execute() should return error for invalid JSON")
	}
}

// TestEngineParseError 测试模板解析错误
func TestEngineParseError(t *testing.T) {
	engine := NewEngine()
	err := engine.Parse("test", `SELECT * FROM users WHERE {invalid`)
	if err == nil {
		t.Error("Parse() should return error for invalid template")
	}
}

// TestNewEngine 测试创建新引擎
func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Error("NewEngine() should not return nil")
	}
}

// TestCRUD_Select 测试 SELECT 查询的各种场景
func TestCRUD_Select(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		// 基础查询
		{
			name:       "SELECT-简单查询全部",
			template:   `SELECT * FROM users`,
			paramsJSON: `{}`,
			wantSQL:    `SELECT * FROM users`,
			wantArgs:   []interface{}{},
		},
		{
			name:       "SELECT-指定字段查询",
			template:   `SELECT id, name, email, created_at FROM users WHERE {expr . "id" "=" "params.id"}`,
			paramsJSON: `{"params": {"id": 1}}`,
			wantSQL:    `SELECT id, name, email, created_at FROM users WHERE id = ?`,
			wantArgs:   []interface{}{float64(1)},
		},
		{
			name:       "SELECT-多条件AND查询",
			template:   `SELECT * FROM users WHERE {and . "and" (expr . "status" "=" "params.status") (expr . "role" "=" "params.role") (expr . "age" ">=" "params.min_age")}`,
			paramsJSON: `{"params": {"status": "active", "role": "admin", "min_age": 18}}`,
			wantSQL:    `SELECT * FROM users WHERE (status = ? and role = ? and age >= ?)`,
			wantArgs:   []interface{}{"active", "admin", float64(18)},
		},
		{
			name:       "SELECT-多条件OR查询",
			template:   `SELECT * FROM users WHERE {or . (expr . "name" "LIKE" "params.keyword") (expr . "email" "LIKE" "params.keyword") (expr . "phone" "LIKE" "params.keyword")}`,
			paramsJSON: `{"params": {"keyword": "%test%"}}`,
			wantSQL:    `SELECT * FROM users WHERE (name LIKE ? or email LIKE ? or phone LIKE ?)`,
			wantArgs:   []interface{}{"%test%", "%test%", "%test%"},
		},
		// IN 查询
		{
			name:       "SELECT-IN查询",
			template:   `SELECT * FROM users WHERE {expr . "id" "IN" "params.ids"}`,
			paramsJSON: `{"params": {"ids": [1, 2, 3, 4, 5]}}`,
			wantSQL:    `SELECT * FROM users WHERE id IN (?, ?, ?, ?, ?)`,
			wantArgs:   []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
		},
		{
			name:       "SELECT-NOT IN查询",
			template:   `SELECT * FROM users WHERE {expr . "status" "NOT IN" "params.exclude_status"}`,
			paramsJSON: `{"params": {"exclude_status": ["deleted", "banned"]}}`,
			wantSQL:    `SELECT * FROM users WHERE status NOT IN (?, ?)`,
			wantArgs:   []interface{}{"deleted", "banned"},
		},
		// BETWEEN 查询
		{
			name:       "SELECT-BETWEEN查询",
			template:   `SELECT * FROM orders WHERE {expr . "amount" "BETWEEN" "params.amount_range"}`,
			paramsJSON: `{"params": {"amount_range": [100, 500]}}`,
			wantSQL:    `SELECT * FROM orders WHERE amount BETWEEN ? AND ?`,
			wantArgs:   []interface{}{float64(100), float64(500)},
		},
		{
			name:       "SELECT-日期BETWEEN查询",
			template:   `SELECT * FROM orders WHERE {expr . "created_at" "BETWEEN" "params.date_range"}`,
			paramsJSON: `{"params": {"date_range": ["2024-01-01", "2024-12-31"]}}`,
			wantSQL:    `SELECT * FROM orders WHERE created_at BETWEEN ? AND ?`,
			wantArgs:   []interface{}{"2024-01-01", "2024-12-31"},
		},
		// LIKE 查询
		{
			name:       "SELECT-LIKE前缀匹配",
			template:   `SELECT * FROM users WHERE {expr . "name" "LIKE" "params.name_prefix"}`,
			paramsJSON: `{"params": {"name_prefix": "张%"}}`,
			wantSQL:    `SELECT * FROM users WHERE name LIKE ?`,
			wantArgs:   []interface{}{"张%"},
		},
		{
			name:       "SELECT-LIKE后缀匹配",
			template:   `SELECT * FROM users WHERE {expr . "email" "LIKE" "params.email_suffix"}`,
			paramsJSON: `{"params": {"email_suffix": "%@gmail.com"}}`,
			wantSQL:    `SELECT * FROM users WHERE email LIKE ?`,
			wantArgs:   []interface{}{"%@gmail.com"},
		},
		{
			name:       "SELECT-LIKE模糊匹配",
			template:   `SELECT * FROM products WHERE {expr . "description" "LIKE" "params.keyword"}`,
			paramsJSON: `{"params": {"keyword": "%手机%"}}`,
			wantSQL:    `SELECT * FROM products WHERE description LIKE ?`,
			wantArgs:   []interface{}{"%手机%"},
		},
		// 比较操作符
		{
			name:       "SELECT-大于查询",
			template:   `SELECT * FROM products WHERE {expr . "price" ">" "params.min_price"}`,
			paramsJSON: `{"params": {"min_price": 100}}`,
			wantSQL:    `SELECT * FROM products WHERE price > ?`,
			wantArgs:   []interface{}{float64(100)},
		},
		{
			name:       "SELECT-小于查询",
			template:   `SELECT * FROM products WHERE {expr . "stock" "<" "params.max_stock"}`,
			paramsJSON: `{"params": {"max_stock": 10}}`,
			wantSQL:    `SELECT * FROM products WHERE stock < ?`,
			wantArgs:   []interface{}{float64(10)},
		},
		{
			name:       "SELECT-大于等于查询",
			template:   `SELECT * FROM orders WHERE {expr . "total" ">=" "params.min_total"}`,
			paramsJSON: `{"params": {"min_total": 1000}}`,
			wantSQL:    `SELECT * FROM orders WHERE total >= ?`,
			wantArgs:   []interface{}{float64(1000)},
		},
		{
			name:       "SELECT-小于等于查询",
			template:   `SELECT * FROM orders WHERE {expr . "discount" "<=" "params.max_discount"}`,
			paramsJSON: `{"params": {"max_discount": 50}}`,
			wantSQL:    `SELECT * FROM orders WHERE discount <= ?`,
			wantArgs:   []interface{}{float64(50)},
		},
		{
			name:       "SELECT-不等于查询",
			template:   `SELECT * FROM users WHERE {expr . "status" "!=" "params.exclude_status"}`,
			paramsJSON: `{"params": {"exclude_status": "deleted"}}`,
			wantSQL:    `SELECT * FROM users WHERE status != ?`,
			wantArgs:   []interface{}{"deleted"},
		},
		// JOIN 查询
		{
			name: "SELECT-INNER JOIN查询",
			template: `SELECT u.id, u.name, o.order_no, o.amount 
FROM users u 
INNER JOIN orders o ON u.id = o.user_id 
WHERE {expr . "u.status" "=" "params.user_status"} AND {expr . "o.status" "=" "params.order_status"}`,
			paramsJSON: `{"params": {"user_status": "active", "order_status": "paid"}}`,
			wantSQL:    `SELECT u.id, u.name, o.order_no, o.amount FROM users u INNER JOIN orders o ON u.id = o.user_id WHERE u.status = ? AND o.status = ?`,
			wantArgs:   []interface{}{"active", "paid"},
		},
		{
			name: "SELECT-LEFT JOIN查询",
			template: `SELECT u.*, p.name as profile_name 
FROM users u 
LEFT JOIN profiles p ON u.id = p.user_id 
WHERE {expr . "u.id" "=" "params.user_id"}`,
			paramsJSON: `{"params": {"user_id": 100}}`,
			wantSQL:    `SELECT u.*, p.name as profile_name FROM users u LEFT JOIN profiles p ON u.id = p.user_id WHERE u.id = ?`,
			wantArgs:   []interface{}{float64(100)},
		},
		// 分页查询
		{
			name:       "SELECT-分页查询",
			template:   `SELECT * FROM users WHERE {expr . "status" "=" "params.status"} LIMIT {val . "params.offset"}, {val . "params.limit"}`,
			paramsJSON: `{"params": {"status": "active", "offset": 0, "limit": 10}}`,
			wantSQL:    `SELECT * FROM users WHERE status = ? LIMIT ?, ?`,
			wantArgs:   []interface{}{"active", float64(0), float64(10)},
		},
		// 排序查询
		{
			name:       "SELECT-带排序查询",
			template:   `SELECT * FROM products WHERE {expr . "category_id" "=" "params.category_id"} ORDER BY created_at DESC LIMIT {val . "params.limit"}`,
			paramsJSON: `{"params": {"category_id": 5, "limit": 20}}`,
			wantSQL:    `SELECT * FROM products WHERE category_id = ? ORDER BY created_at DESC LIMIT ?`,
			wantArgs:   []interface{}{float64(5), float64(20)},
		},
		// 聚合查询
		{
			name:       "SELECT-COUNT聚合",
			template:   `SELECT COUNT(*) as total FROM users WHERE {expr . "status" "=" "params.status"}`,
			paramsJSON: `{"params": {"status": "active"}}`,
			wantSQL:    `SELECT COUNT(*) as total FROM users WHERE status = ?`,
			wantArgs:   []interface{}{"active"},
		},
		{
			name:       "SELECT-SUM聚合",
			template:   `SELECT SUM(amount) as total_amount FROM orders WHERE {expr . "user_id" "=" "params.user_id"} AND {expr . "status" "=" "params.status"}`,
			paramsJSON: `{"params": {"user_id": 1, "status": "completed"}}`,
			wantSQL:    `SELECT SUM(amount) as total_amount FROM orders WHERE user_id = ? AND status = ?`,
			wantArgs:   []interface{}{float64(1), "completed"},
		},
		// GROUP BY 查询
		{
			name:       "SELECT-GROUP BY查询",
			template:   `SELECT category_id, COUNT(*) as count FROM products WHERE {expr . "status" "=" "params.status"} GROUP BY category_id`,
			paramsJSON: `{"params": {"status": "active"}}`,
			wantSQL:    `SELECT category_id, COUNT(*) as count FROM products WHERE status = ? GROUP BY category_id`,
			wantArgs:   []interface{}{"active"},
		},
		{
			name:       "SELECT-GROUP BY HAVING查询",
			template:   `SELECT user_id, SUM(amount) as total FROM orders WHERE {expr . "status" "=" "params.status"} GROUP BY user_id HAVING total > {val . "params.min_total"}`,
			paramsJSON: `{"params": {"status": "completed", "min_total": 1000}}`,
			wantSQL:    `SELECT user_id, SUM(amount) as total FROM orders WHERE status = ? GROUP BY user_id HAVING total > ?`,
			wantArgs:   []interface{}{"completed", float64(1000)},
		},
		// 子查询
		{
			name:       "SELECT-子查询IN",
			template:   `SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE {expr . "status" "=" "params.order_status"})`,
			paramsJSON: `{"params": {"order_status": "completed"}}`,
			wantSQL:    `SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE status = ?)`,
			wantArgs:   []interface{}{"completed"},
		},
		// 条件裁剪
		{
			name: "SELECT-条件裁剪全部存在",
			template: `SELECT * FROM users WHERE 1=1 
{if not (isEmpty (getValue . "params.name"))} AND {expr . "name" "=" "params.name"}{end}
{if not (isEmpty (getValue . "params.status"))} AND {expr . "status" "=" "params.status"}{end}
{if not (isEmpty (getValue . "params.role"))} AND {expr . "role" "=" "params.role"}{end}`,
			paramsJSON: `{"params": {"name": "张三", "status": "active", "role": "admin"}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1 AND name = ? AND status = ? AND role = ?`,
			wantArgs:   []interface{}{"张三", "active", "admin"},
		},
		{
			name: "SELECT-条件裁剪部分存在",
			template: `SELECT * FROM users WHERE 1=1 
{if not (isEmpty (getValue . "params.name"))} AND {expr . "name" "=" "params.name"}{end}
{if not (isEmpty (getValue . "params.status"))} AND {expr . "status" "=" "params.status"}{end}
{if not (isEmpty (getValue . "params.role"))} AND {expr . "role" "=" "params.role"}{end}`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1 AND name = ?`,
			wantArgs:   []interface{}{"张三"},
		},
		{
			name: "SELECT-条件裁剪全部为空",
			template: `SELECT * FROM users WHERE 1=1 
{if not (isEmpty (getValue . "params.name"))} AND {expr . "name" "=" "params.name"}{end}
{if not (isEmpty (getValue . "params.status"))} AND {expr . "status" "=" "params.status"}{end}`,
			paramsJSON: `{"params": {}}`,
			wantSQL:    `SELECT * FROM users WHERE 1=1`,
			wantArgs:   []interface{}{},
		},
		// 复杂嵌套查询
		{
			name: "SELECT-复杂嵌套AND+OR",
			template: `SELECT * FROM products WHERE {and . "and" 
(expr . "status" "=" "params.status")
(expr . "category_id" "IN" "params.category_ids")
(or . 
	(expr . "name" "LIKE" "params.keyword")
	(expr . "description" "LIKE" "params.keyword")
)
(expr . "price" "BETWEEN" "params.price_range")
}`,
			paramsJSON: `{"params": {"status": "active", "category_ids": [1,2,3], "keyword": "%手机%", "price_range": [1000, 5000]}}`,
			wantSQL:    `SELECT * FROM products WHERE (status = ? and category_id IN (?, ?, ?) and (name LIKE ? or description LIKE ?) and price BETWEEN ? AND ?)`,
			wantArgs:   []interface{}{"active", float64(1), float64(2), float64(3), "%手机%", "%手机%", float64(1000), float64(5000)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestCRUD_Insert 测试 INSERT 插入的各种场景
func TestCRUD_Insert(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		// 单行插入
		{
			name:       "INSERT-单字段插入",
			template:   `INSERT INTO users (name) VALUES ({val . "params.name"})`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `INSERT INTO users (name) VALUES (?)`,
			wantArgs:   []interface{}{"张三"},
		},
		{
			name:       "INSERT-多字段插入",
			template:   `INSERT INTO users (name, email, age, status) VALUES ({val . "params.name"}, {val . "params.email"}, {val . "params.age"}, {val . "params.status"})`,
			paramsJSON: `{"params": {"name": "张三", "email": "zhangsan@example.com", "age": 25, "status": "active"}}`,
			wantSQL:    `INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			wantArgs:   []interface{}{"张三", "zhangsan@example.com", float64(25), "active"},
		},
		{
			name:       "INSERT-带NULL值插入",
			template:   `INSERT INTO users (name, nickname, phone) VALUES ({val . "params.name"}, {val . "params.nickname"}, {val . "params.phone"})`,
			paramsJSON: `{"params": {"name": "张三"}}`,
			wantSQL:    `INSERT INTO users (name, nickname, phone) VALUES (?, ?, ?)`,
			wantArgs:   []interface{}{"张三", nil, nil},
		},
		{
			name:       "INSERT-布尔值插入",
			template:   `INSERT INTO users (name, is_vip, is_verified) VALUES ({val . "params.name"}, {val . "params.is_vip"}, {val . "params.is_verified"})`,
			paramsJSON: `{"params": {"name": "张三", "is_vip": true, "is_verified": false}}`,
			wantSQL:    `INSERT INTO users (name, is_vip, is_verified) VALUES (?, ?, ?)`,
			wantArgs:   []interface{}{"张三", true, false},
		},
		{
			name:       "INSERT-嵌套对象字段插入",
			template:   `INSERT INTO users (name, city, province) VALUES ({val . "params.name"}, {val . "params.address.city"}, {val . "params.address.province"})`,
			paramsJSON: `{"params": {"name": "张三", "address": {"city": "北京", "province": "北京市"}}}`,
			wantSQL:    `INSERT INTO users (name, city, province) VALUES (?, ?, ?)`,
			wantArgs:   []interface{}{"张三", "北京", "北京市"},
		},
		// 批量插入
		{
			name:       "INSERT-批量插入单字段",
			template:   `INSERT INTO tags (name) VALUES {range $i, $_ := (getValue . "params.tags")}{if $i}, {end}({val $ (printf "params.tags.%d" $i)}){end}`,
			paramsJSON: `{"params": {"tags": ["标签1", "标签2", "标签3"]}}`,
			wantSQL:    `INSERT INTO tags (name) VALUES (?), (?), (?)`,
			wantArgs:   []interface{}{"标签1", "标签2", "标签3"},
		},
		{
			name:     "INSERT-批量插入多字段",
			template: `INSERT INTO users (name, age, status) VALUES {range $i, $_ := (getValue . "params.users")}{if $i}, {end}({val $ (printf "params.users.%d.name" $i)}, {val $ (printf "params.users.%d.age" $i)}, {val $ (printf "params.users.%d.status" $i)}){end}`,
			paramsJSON: `{"params": {"users": [
				{"name": "张三", "age": 25, "status": "active"},
				{"name": "李四", "age": 30, "status": "active"},
				{"name": "王五", "age": 28, "status": "inactive"}
			]}}`,
			wantSQL:  `INSERT INTO users (name, age, status) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)`,
			wantArgs: []interface{}{"张三", float64(25), "active", "李四", float64(30), "active", "王五", float64(28), "inactive"},
		},
		// INSERT ... SELECT
		{
			name:       "INSERT-SELECT子查询插入",
			template:   `INSERT INTO user_archives (user_id, name, created_at) SELECT id, name, created_at FROM users WHERE {expr . "status" "=" "params.status"}`,
			paramsJSON: `{"params": {"status": "deleted"}}`,
			wantSQL:    `INSERT INTO user_archives (user_id, name, created_at) SELECT id, name, created_at FROM users WHERE status = ?`,
			wantArgs:   []interface{}{"deleted"},
		},
		// INSERT ... ON DUPLICATE KEY UPDATE (MySQL)
		{
			name:       "INSERT-ON DUPLICATE KEY UPDATE",
			template:   `INSERT INTO user_stats (user_id, login_count, last_login) VALUES ({val . "params.user_id"}, {val . "params.login_count"}, {val . "params.last_login"}) ON DUPLICATE KEY UPDATE login_count = login_count + 1, last_login = {val . "params.last_login"}`,
			paramsJSON: `{"params": {"user_id": 1, "login_count": 1, "last_login": "2024-01-15 10:00:00"}}`,
			wantSQL:    `INSERT INTO user_stats (user_id, login_count, last_login) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE login_count = login_count + 1, last_login = ?`,
			wantArgs:   []interface{}{float64(1), float64(1), "2024-01-15 10:00:00", "2024-01-15 10:00:00"},
		},
		// INSERT IGNORE
		{
			name:       "INSERT-IGNORE插入",
			template:   `INSERT IGNORE INTO users (id, name, email) VALUES ({val . "params.id"}, {val . "params.name"}, {val . "params.email"})`,
			paramsJSON: `{"params": {"id": 1, "name": "张三", "email": "test@example.com"}}`,
			wantSQL:    `INSERT IGNORE INTO users (id, name, email) VALUES (?, ?, ?)`,
			wantArgs:   []interface{}{float64(1), "张三", "test@example.com"},
		},
		// REPLACE INTO
		{
			name:       "REPLACE-替换插入",
			template:   `REPLACE INTO users (id, name, email) VALUES ({val . "params.id"}, {val . "params.name"}, {val . "params.email"})`,
			paramsJSON: `{"params": {"id": 1, "name": "张三更新", "email": "new@example.com"}}`,
			wantSQL:    `REPLACE INTO users (id, name, email) VALUES (?, ?, ?)`,
			wantArgs:   []interface{}{float64(1), "张三更新", "new@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestCRUD_Update 测试 UPDATE 更新的各种场景
func TestCRUD_Update(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		// 单字段更新
		{
			name:       "UPDATE-单字段更新",
			template:   `UPDATE users SET name = {val . "params.name"} WHERE {expr . "id" "=" "params.id"}`,
			paramsJSON: `{"params": {"name": "张三新", "id": 1}}`,
			wantSQL:    `UPDATE users SET name = ? WHERE id = ?`,
			wantArgs:   []interface{}{"张三新", float64(1)},
		},
		// 多字段更新
		{
			name:       "UPDATE-多字段更新",
			template:   `UPDATE users SET name = {val . "params.name"}, email = {val . "params.email"}, age = {val . "params.age"}, updated_at = {val . "params.updated_at"} WHERE {expr . "id" "=" "params.id"}`,
			paramsJSON: `{"params": {"name": "张三新", "email": "new@example.com", "age": 26, "updated_at": "2024-01-15 10:00:00", "id": 1}}`,
			wantSQL:    `UPDATE users SET name = ?, email = ?, age = ?, updated_at = ? WHERE id = ?`,
			wantArgs:   []interface{}{"张三新", "new@example.com", float64(26), "2024-01-15 10:00:00", float64(1)},
		},
		// 条件更新
		{
			name:       "UPDATE-多条件更新",
			template:   `UPDATE users SET status = {val . "params.new_status"} WHERE {and . "and" (expr . "status" "=" "params.old_status") (expr . "role" "=" "params.role")}`,
			paramsJSON: `{"params": {"new_status": "inactive", "old_status": "active", "role": "user"}}`,
			wantSQL:    `UPDATE users SET status = ? WHERE (status = ? and role = ?)`,
			wantArgs:   []interface{}{"inactive", "active", "user"},
		},
		{
			name:       "UPDATE-IN条件更新",
			template:   `UPDATE users SET status = {val . "params.status"} WHERE {expr . "id" "IN" "params.ids"}`,
			paramsJSON: `{"params": {"status": "banned", "ids": [1, 2, 3, 4, 5]}}`,
			wantSQL:    `UPDATE users SET status = ? WHERE id IN (?, ?, ?, ?, ?)`,
			wantArgs:   []interface{}{"banned", float64(1), float64(2), float64(3), float64(4), float64(5)},
		},
		// 自增更新
		{
			name:       "UPDATE-自增更新",
			template:   `UPDATE products SET stock = stock - {val . "params.quantity"}, sales = sales + {val . "params.quantity"} WHERE {expr . "id" "=" "params.product_id"} AND stock >= {val . "params.quantity"}`,
			paramsJSON: `{"params": {"quantity": 5, "product_id": 100}}`,
			wantSQL:    `UPDATE products SET stock = stock - ?, sales = sales + ? WHERE id = ? AND stock >= ?`,
			wantArgs:   []interface{}{float64(5), float64(5), float64(100), float64(5)},
		},
		// 条件裁剪更新
		{
			name: "UPDATE-动态字段更新",
			template: `UPDATE users SET updated_at = NOW()
{if not (isEmpty (getValue . "params.name"))}, name = {val . "params.name"}{end}
{if not (isEmpty (getValue . "params.email"))}, email = {val . "params.email"}{end}
{if not (isEmpty (getValue . "params.phone"))}, phone = {val . "params.phone"}{end}
WHERE {expr . "id" "=" "params.id"}`,
			paramsJSON: `{"params": {"name": "张三新", "email": "new@example.com", "id": 1}}`,
			wantSQL:    `UPDATE users SET updated_at = NOW() , name = ? , email = ? WHERE id = ?`,
			wantArgs:   []interface{}{"张三新", "new@example.com", float64(1)},
		},
		// JOIN 更新
		{
			name:       "UPDATE-JOIN更新",
			template:   `UPDATE orders o INNER JOIN users u ON o.user_id = u.id SET o.status = {val . "params.order_status"} WHERE {expr . "u.status" "=" "params.user_status"}`,
			paramsJSON: `{"params": {"order_status": "cancelled", "user_status": "banned"}}`,
			wantSQL:    `UPDATE orders o INNER JOIN users u ON o.user_id = u.id SET o.status = ? WHERE u.status = ?`,
			wantArgs:   []interface{}{"cancelled", "banned"},
		},
		// CASE WHEN 更新
		{
			name:       "UPDATE-CASE WHEN批量状态更新",
			template:   `UPDATE orders SET status = CASE WHEN amount > {val . "params.threshold"} THEN {val . "params.high_status"} ELSE {val . "params.low_status"} END WHERE {expr . "id" "IN" "params.order_ids"}`,
			paramsJSON: `{"params": {"threshold": 1000, "high_status": "priority", "low_status": "normal", "order_ids": [1, 2, 3]}}`,
			wantSQL:    `UPDATE orders SET status = CASE WHEN amount > ? THEN ? ELSE ? END WHERE id IN (?, ?, ?)`,
			wantArgs:   []interface{}{float64(1000), "priority", "normal", float64(1), float64(2), float64(3)},
		},
		// LIMIT 更新
		{
			name:       "UPDATE-LIMIT更新",
			template:   `UPDATE users SET status = {val . "params.status"} WHERE {expr . "last_login" "<" "params.inactive_before"} ORDER BY last_login ASC LIMIT {val . "params.limit"}`,
			paramsJSON: `{"params": {"status": "inactive", "inactive_before": "2023-01-01", "limit": 100}}`,
			wantSQL:    `UPDATE users SET status = ? WHERE last_login < ? ORDER BY last_login ASC LIMIT ?`,
			wantArgs:   []interface{}{"inactive", "2023-01-01", float64(100)},
		},
		// 批量动态更新
		{
			name:     "UPDATE-批量动态字段更新",
			template: `UPDATE users SET {range $i, $_ := (getValue . "params.updates")}{if $i}, {end}{val $ (printf "params.updates.%d.field" $i)} = {val $ (printf "params.updates.%d.value" $i)}{end} WHERE {expr . "id" "=" "params.id"}`,
			paramsJSON: `{"params": {"updates": [
				{"field": "name", "value": "新名字"},
				{"field": "email", "value": "new@test.com"},
				{"field": "status", "value": "active"}
			], "id": 1}}`,
			wantSQL:  `UPDATE users SET ? = ?, ? = ?, ? = ? WHERE id = ?`,
			wantArgs: []interface{}{"name", "新名字", "email", "new@test.com", "status", "active", float64(1)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestCRUD_Delete 测试 DELETE 删除的各种场景
func TestCRUD_Delete(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		// 简单删除
		{
			name:       "DELETE-按ID删除",
			template:   `DELETE FROM users WHERE {expr . "id" "=" "params.id"}`,
			paramsJSON: `{"params": {"id": 1}}`,
			wantSQL:    `DELETE FROM users WHERE id = ?`,
			wantArgs:   []interface{}{float64(1)},
		},
		// 多条件删除
		{
			name:       "DELETE-多条件AND删除",
			template:   `DELETE FROM users WHERE {and . "and" (expr . "status" "=" "params.status") (expr . "role" "=" "params.role")}`,
			paramsJSON: `{"params": {"status": "deleted", "role": "user"}}`,
			wantSQL:    `DELETE FROM users WHERE (status = ? and role = ?)`,
			wantArgs:   []interface{}{"deleted", "user"},
		},
		{
			name:       "DELETE-多条件OR删除",
			template:   `DELETE FROM logs WHERE {or . (expr . "level" "=" "params.level1") (expr . "level" "=" "params.level2")}`,
			paramsJSON: `{"params": {"level1": "debug", "level2": "trace"}}`,
			wantSQL:    `DELETE FROM logs WHERE (level = ? or level = ?)`,
			wantArgs:   []interface{}{"debug", "trace"},
		},
		// IN 删除
		{
			name:       "DELETE-IN条件删除",
			template:   `DELETE FROM users WHERE {expr . "id" "IN" "params.ids"}`,
			paramsJSON: `{"params": {"ids": [1, 2, 3, 4, 5]}}`,
			wantSQL:    `DELETE FROM users WHERE id IN (?, ?, ?, ?, ?)`,
			wantArgs:   []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
		},
		{
			name:       "DELETE-NOT IN条件删除",
			template:   `DELETE FROM sessions WHERE {expr . "user_id" "NOT IN" "params.active_user_ids"}`,
			paramsJSON: `{"params": {"active_user_ids": [1, 2, 3]}}`,
			wantSQL:    `DELETE FROM sessions WHERE user_id NOT IN (?, ?, ?)`,
			wantArgs:   []interface{}{float64(1), float64(2), float64(3)},
		},
		// 范围删除
		{
			name:       "DELETE-小于条件删除",
			template:   `DELETE FROM logs WHERE {expr . "created_at" "<" "params.before_date"}`,
			paramsJSON: `{"params": {"before_date": "2023-01-01"}}`,
			wantSQL:    `DELETE FROM logs WHERE created_at < ?`,
			wantArgs:   []interface{}{"2023-01-01"},
		},
		{
			name:       "DELETE-BETWEEN范围删除",
			template:   `DELETE FROM logs WHERE {expr . "created_at" "BETWEEN" "params.date_range"}`,
			paramsJSON: `{"params": {"date_range": ["2023-01-01", "2023-06-30"]}}`,
			wantSQL:    `DELETE FROM logs WHERE created_at BETWEEN ? AND ?`,
			wantArgs:   []interface{}{"2023-01-01", "2023-06-30"},
		},
		// LIKE 删除
		{
			name:       "DELETE-LIKE条件删除",
			template:   `DELETE FROM temp_files WHERE {expr . "filename" "LIKE" "params.pattern"}`,
			paramsJSON: `{"params": {"pattern": "%.tmp"}}`,
			wantSQL:    `DELETE FROM temp_files WHERE filename LIKE ?`,
			wantArgs:   []interface{}{"%.tmp"},
		},
		// LIMIT 删除
		{
			name:       "DELETE-LIMIT删除",
			template:   `DELETE FROM logs WHERE {expr . "level" "=" "params.level"} ORDER BY created_at ASC LIMIT {val . "params.limit"}`,
			paramsJSON: `{"params": {"level": "debug", "limit": 1000}}`,
			wantSQL:    `DELETE FROM logs WHERE level = ? ORDER BY created_at ASC LIMIT ?`,
			wantArgs:   []interface{}{"debug", float64(1000)},
		},
		// 子查询删除
		{
			name:       "DELETE-子查询删除",
			template:   `DELETE FROM orders WHERE user_id IN (SELECT id FROM users WHERE {expr . "status" "=" "params.user_status"})`,
			paramsJSON: `{"params": {"user_status": "deleted"}}`,
			wantSQL:    `DELETE FROM orders WHERE user_id IN (SELECT id FROM users WHERE status = ?)`,
			wantArgs:   []interface{}{"deleted"},
		},
		// JOIN 删除 (MySQL)
		{
			name:       "DELETE-JOIN删除",
			template:   `DELETE o FROM orders o INNER JOIN users u ON o.user_id = u.id WHERE {expr . "u.status" "=" "params.user_status"}`,
			paramsJSON: `{"params": {"user_status": "banned"}}`,
			wantSQL:    `DELETE o FROM orders o INNER JOIN users u ON o.user_id = u.id WHERE u.status = ?`,
			wantArgs:   []interface{}{"banned"},
		},
		// 多表删除
		{
			name:       "DELETE-多表删除",
			template:   `DELETE u, p FROM users u LEFT JOIN profiles p ON u.id = p.user_id WHERE {expr . "u.id" "=" "params.user_id"}`,
			paramsJSON: `{"params": {"user_id": 100}}`,
			wantSQL:    `DELETE u, p FROM users u LEFT JOIN profiles p ON u.id = p.user_id WHERE u.id = ?`,
			wantArgs:   []interface{}{float64(100)},
		},
		// 条件裁剪删除
		{
			name: "DELETE-动态条件删除",
			template: `DELETE FROM logs WHERE 1=1
{if not (isEmpty (getValue . "params.level"))} AND {expr . "level" "=" "params.level"}{end}
{if not (isEmpty (getValue . "params.before_date"))} AND {expr . "created_at" "<" "params.before_date"}{end}
{if not (isEmpty (getValue . "params.user_id"))} AND {expr . "user_id" "=" "params.user_id"}{end}`,
			paramsJSON: `{"params": {"level": "debug", "before_date": "2023-01-01"}}`,
			wantSQL:    `DELETE FROM logs WHERE 1=1 AND level = ? AND created_at < ?`,
			wantArgs:   []interface{}{"debug", "2023-01-01"},
		},
		// 复杂条件删除
		{
			name: "DELETE-复杂嵌套条件删除",
			template: `DELETE FROM notifications WHERE {and . "and"
(expr . "status" "=" "params.status")
(or .
	(expr . "type" "IN" "params.types")
	(expr . "created_at" "<" "params.expire_date")
)
}`,
			paramsJSON: `{"params": {"status": "read", "types": ["promotion", "system"], "expire_date": "2023-06-01"}}`,
			wantSQL:    `DELETE FROM notifications WHERE (status = ? and (type IN (?, ?) or created_at < ?))`,
			wantArgs:   []interface{}{"read", "promotion", "system", "2023-06-01"},
		},
		// 批量循环删除条件
		{
			name:       "DELETE-range循环条件删除",
			template:   `DELETE FROM user_roles WHERE {range $i, $_ := (getValue . "params.roles")}{if $i} OR {end}({expr $ "user_id" "=" "params.user_id"} AND {expr $ "role_id" "=" (printf "params.roles.%d" $i)}){end}`,
			paramsJSON: `{"params": {"user_id": 1, "roles": [10, 20, 30]}}`,
			wantSQL:    `DELETE FROM user_roles WHERE (user_id = ? AND role_id = ?) OR (user_id = ? AND role_id = ?) OR (user_id = ? AND role_id = ?)`,
			wantArgs:   []interface{}{float64(1), float64(10), float64(1), float64(20), float64(1), float64(30)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}

// TestCRUD_Complex 测试复杂的综合场景
func TestCRUD_Complex(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		paramsJSON string
		wantSQL    string
		wantArgs   []interface{}
	}{
		// 电商订单查询场景
		{
			name: "复杂场景-电商订单多条件查询",
			template: `SELECT o.*, u.name as user_name, p.name as product_name
FROM orders o
INNER JOIN users u ON o.user_id = u.id
INNER JOIN order_items oi ON o.id = oi.order_id
INNER JOIN products p ON oi.product_id = p.id
WHERE 1=1
{if not (isEmpty (getValue . "params.order_no"))} AND {expr . "o.order_no" "=" "params.order_no"}{end}
{if not (isEmpty (getValue . "params.user_id"))} AND {expr . "o.user_id" "=" "params.user_id"}{end}
{if not (isEmpty (getValue . "params.status"))} AND {expr . "o.status" "IN" "params.status"}{end}
{if not (isEmpty (getValue . "params.date_range"))} AND {expr . "o.created_at" "BETWEEN" "params.date_range"}{end}
{if not (isEmpty (getValue . "params.min_amount"))} AND {expr . "o.total_amount" ">=" "params.min_amount"}{end}
{if not (isEmpty (getValue . "params.keyword"))} AND {or . (expr . "o.order_no" "LIKE" "params.keyword") (expr . "u.name" "LIKE" "params.keyword")}{end}
ORDER BY o.created_at DESC
LIMIT {val . "params.offset"}, {val . "params.limit"}`,
			paramsJSON: `{"params": {"status": ["paid", "shipped"], "date_range": ["2024-01-01", "2024-12-31"], "min_amount": 100, "keyword": "%张%", "offset": 0, "limit": 20}}`,
			wantSQL:    `SELECT o.*, u.name as user_name, p.name as product_name FROM orders o INNER JOIN users u ON o.user_id = u.id INNER JOIN order_items oi ON o.id = oi.order_id INNER JOIN products p ON oi.product_id = p.id WHERE 1=1 AND o.status IN (?, ?) AND o.created_at BETWEEN ? AND ? AND o.total_amount >= ? AND (o.order_no LIKE ? or u.name LIKE ?) ORDER BY o.created_at DESC LIMIT ?, ?`,
			wantArgs:   []interface{}{"paid", "shipped", "2024-01-01", "2024-12-31", float64(100), "%张%", "%张%", float64(0), float64(20)},
		},
		// 库存扣减事务场景
		{
			name: "复杂场景-库存扣减批量更新",
			template: `UPDATE products SET 
stock = stock - CASE id {range $i, $_ := (getValue . "params.items")} WHEN {val $ (printf "params.items.%d.product_id" $i)} THEN {val $ (printf "params.items.%d.quantity" $i)}{end} END,
sales = sales + CASE id {range $i, $_ := (getValue . "params.items")} WHEN {val $ (printf "params.items.%d.product_id" $i)} THEN {val $ (printf "params.items.%d.quantity" $i)}{end} END,
updated_at = {val . "params.updated_at"}
WHERE {expr . "id" "IN" "params.product_ids"}`,
			paramsJSON: `{"params": {"items": [{"product_id": 1, "quantity": 2}, {"product_id": 2, "quantity": 3}], "product_ids": [1, 2], "updated_at": "2024-01-15 10:00:00"}}`,
			wantSQL:    `UPDATE products SET stock = stock - CASE id WHEN ? THEN ? WHEN ? THEN ? END, sales = sales + CASE id WHEN ? THEN ? WHEN ? THEN ? END, updated_at = ? WHERE id IN (?, ?)`,
			wantArgs:   []interface{}{float64(1), float64(2), float64(2), float64(3), float64(1), float64(2), float64(2), float64(3), "2024-01-15 10:00:00", float64(1), float64(2)},
		},
		// 用户权限检查场景
		{
			name: "复杂场景-用户权限多级检查",
			template: `SELECT COUNT(*) as has_permission FROM user_permissions up
INNER JOIN roles r ON up.role_id = r.id
WHERE {and . "and"
(expr . "up.user_id" "=" "params.user_id")
(or .
	(expr . "up.permission_code" "=" "params.permission")
	(expr . "r.is_super_admin" "=" "params.is_super_admin")
)
(expr . "up.status" "=" "params.status")
}`,
			paramsJSON: `{"params": {"user_id": 1, "permission": "user:create", "is_super_admin": true, "status": "active"}}`,
			wantSQL:    `SELECT COUNT(*) as has_permission FROM user_permissions up INNER JOIN roles r ON up.role_id = r.id WHERE (up.user_id = ? and (up.permission_code = ? or r.is_super_admin = ?) and up.status = ?)`,
			wantArgs:   []interface{}{float64(1), "user:create", true, "active"},
		},
		// 报表统计场景
		{
			name: "复杂场景-销售报表统计",
			template: `SELECT 
DATE_FORMAT(o.created_at, '%Y-%m') as month,
COUNT(DISTINCT o.id) as order_count,
COUNT(DISTINCT o.user_id) as user_count,
SUM(o.total_amount) as total_sales,
AVG(o.total_amount) as avg_order_amount
FROM orders o
WHERE {and . "and"
(expr . "o.status" "IN" "params.status")
(expr . "o.created_at" "BETWEEN" "params.date_range")
(expr . "o.category_id" "=" "params.category_id")
}
GROUP BY DATE_FORMAT(o.created_at, '%Y-%m')
HAVING total_sales > {val . "params.min_sales"}
ORDER BY month DESC`,
			paramsJSON: `{"params": {"status": ["completed", "refunded"], "date_range": ["2024-01-01", "2024-12-31"], "category_id": 5, "min_sales": 10000}}`,
			wantSQL:    `SELECT DATE_FORMAT(o.created_at, '%Y-%m') as month, COUNT(DISTINCT o.id) as order_count, COUNT(DISTINCT o.user_id) as user_count, SUM(o.total_amount) as total_sales, AVG(o.total_amount) as avg_order_amount FROM orders o WHERE (o.status IN (?, ?) and o.created_at BETWEEN ? AND ? and o.category_id = ?) GROUP BY DATE_FORMAT(o.created_at, '%Y-%m') HAVING total_sales > ? ORDER BY month DESC`,
			wantArgs:   []interface{}{"completed", "refunded", "2024-01-01", "2024-12-31", float64(5), float64(10000)},
		},
		// 批量订单创建场景
		{
			name: "复杂场景-批量订单明细插入",
			template: `INSERT INTO order_items (order_id, product_id, quantity, unit_price, total_price, created_at) VALUES 
{range $i, $_ := (getValue . "params.items")}{if $i}, {end}({val $ "params.order_id"}, {val $ (printf "params.items.%d.product_id" $i)}, {val $ (printf "params.items.%d.quantity" $i)}, {val $ (printf "params.items.%d.unit_price" $i)}, {val $ (printf "params.items.%d.total_price" $i)}, {val $ "params.created_at"}){end}`,
			paramsJSON: `{"params": {"order_id": 1001, "items": [{"product_id": 1, "quantity": 2, "unit_price": 99.9, "total_price": 199.8}, {"product_id": 2, "quantity": 1, "unit_price": 199.9, "total_price": 199.9}], "created_at": "2024-01-15 10:00:00"}}`,
			wantSQL:    `INSERT INTO order_items (order_id, product_id, quantity, unit_price, total_price, created_at) VALUES (?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?)`,
			wantArgs:   []interface{}{float64(1001), float64(1), float64(2), 99.9, 199.8, "2024-01-15 10:00:00", float64(1001), float64(2), float64(1), 199.9, 199.9, "2024-01-15 10:00:00"},
		},
		// 软删除与清理场景
		{
			name: "复杂场景-过期数据清理",
			template: `DELETE FROM audit_logs 
WHERE {and . "and"
(expr . "created_at" "<" "params.expire_date")
(expr . "level" "IN" "params.levels")
(or .
	(expr . "status" "=" "params.processed_status")
	(and . "and"
		(expr . "status" "=" "params.failed_status")
		(expr . "retry_count" ">=" "params.max_retry")
	)
)
}
LIMIT {val . "params.batch_size"}`,
			paramsJSON: `{"params": {"expire_date": "2023-01-01", "levels": ["debug", "info"], "processed_status": "processed", "failed_status": "failed", "max_retry": 3, "batch_size": 1000}}`,
			wantSQL:    `DELETE FROM audit_logs WHERE (created_at < ? and level IN (?, ?) and (status = ? or (status = ? and retry_count >= ?))) LIMIT ?`,
			wantArgs:   []interface{}{"2023-01-01", "debug", "info", "processed", "failed", float64(3), float64(1000)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			if err := engine.Parse("test", tt.template); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			result, err := engine.Execute(tt.paramsJSON)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if result.SQL != tt.wantSQL {
				t.Errorf("Execute() SQL = %q, want %q", result.SQL, tt.wantSQL)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("Execute() Args length = %v, want %v", len(result.Args), len(tt.wantArgs))
				return
			}

			for i, wantArg := range tt.wantArgs {
				if !compareValues(result.Args[i], wantArg) {
					t.Errorf("Execute() Args[%d] = %v (%T), want %v (%T)",
						i, result.Args[i], result.Args[i], wantArg, wantArg)
				}
			}
		})
	}
}
