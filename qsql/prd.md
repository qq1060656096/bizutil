
# SQL 占位符引擎 RFC（完整版 + val）

## 一、设计目标

1. **SQL 主体固定，所有条件占位符化**
2. **支持可嵌套组合**：AND / OR / expr / if / for / val
3. **支持循环生成**：for 遍历数组，生成重复条件
4. **支持条件裁剪**：空参数自动忽略，不生成 SQL
5. **支持动态值插入**：val 插入字面量 / 计算值 / 动态字段
6. **输出**：安全的预编译 SQL + args，BI 查询能力

---

## 二、核心占位符定义

### 1️⃣ expr：原子条件

```sql
{{expr "field" "op" "$.params.xxx"}}
```

| 参数           | 说明                      |
| ------------ | ----------------------- |
| field        | 数据库字段                   |
| op           | = / in / > / < / like 等 |
| $.params.xxx | 前端 JSON 参数路径            |

**规则**：

* 参数为空 → expr 不生成
* `in` 自动展开 `(?, ?, ?)`
* 可嵌套在 `and` / `or` / `for` 内

---

### 2️⃣ and / or：逻辑组合

```sql
{{and
    expr1
    expr2
}}
```

→ `(expr1 AND expr2)`

```sql
{{or
    expr1
    expr2
}}
```

→ `(expr1 OR expr2)`

* 可嵌套任意层级
* 自动裁剪空 expr，不生成空括号

---

### 3️⃣ if：条件裁剪

```sql
{{if condition}}
    SQL / 占位符
{{end}}
```

* 条件语法：

  ```text
  $.params.xxx
  $.params.xxx && $.params.yyy
  $.params.xxx || $.params.yyy
  !$.params.xxx
  ```
* 功能：控制整段 SQL 是否渲染
* 可嵌套 expr / and / or / for / val

---

### 4️⃣ for：循环生成

```sql
{{for "$.params.list" as item}}
    SQL / 占位符
{{end}}
```

* 遍历数组生成重复条件
* `item` 可作为循环变量引用
* 可嵌套逻辑组合和 expr / val

---

### 5️⃣ val：字面量 / 内联 SQL

```sql
{{val "$.params.xxx"}}
```

* 直接插入值或表达式到 SQL
* 不生成 `?` 占位符
* 可用于：

  * 常量
  * 动态字段
  * 排序 / 分组 / 计算值
  * 可配合循环、逻辑组合

**示例**：

```sql
SELECT * FROM user
WHERE id = {{val "$.user_id"}}
ORDER BY {{val "$.sort_field"}} {{val "$.sort_order"}}
```

> ⚠️ 安全提醒：val 会直接插入 SQL，必须保证来源可信或转义

---

## 三、典型使用示例

### 1️⃣ 单字段 in 条件

```sql
SELECT * FROM user
WHERE 1=1
{{and
    {{expr "name" "in" "$.params.names"}}
}}
```

```json
{ "names": ["张三","李四"] }
```

生成 SQL：

```sql
WHERE 1=1 AND name IN (?, ?)
```

---

### 2️⃣ OR 条件 + 循环生成

```sql
SELECT * FROM user
WHERE 1=1
{{or
    {{for "$.params.user_ids" as uid}}
        {{expr "user_id" "=" "uid"}}
    {{end}}
}}
```

```json
{ "user_ids": [1,2,3] }
```

生成 SQL：

```sql
WHERE 1=1 AND (user_id = ? OR user_id = ? OR user_id = ?)
```

---

### 3️⃣ 复杂嵌套 + if 自动裁剪

```sql
SELECT * FROM user
WHERE 1=1
{{if $.params.group1 || $.params.group2}}
{{or
    {{if $.params.group1}}
    {{and
        {{expr "user_id" "=" "$.params.uid1"}}
        {{expr "name" "=" "$.params.name1"}}
    }}
    {{end}}

    {{if $.params.group2}}
    {{and
        {{expr "user_id" "=" "$.params.uid2"}}
        {{expr "name" "=" "$.params.name2"}}
    }}
    {{end}}
}}
{{end}}
```

> 空参数自动裁剪，只保留有效逻辑块

---

### 4️⃣ 循环生成多条件 + val

```sql
SELECT * FROM orders
WHERE 1=1
{{for "$.params.items" as item}}
{{and
    {{expr "product_id" "=" "item.id"}}
    {{expr "qty" ">" "item.qty"}}
    {{val "item.extra_condition"}}
}}
{{end}}
```

```json
{
  "items": [
    { "id": 101, "qty": 2, "extra_condition": "AND discount > 0" },
    { "id": 102, "qty": 5, "extra_condition": "AND discount > 5" }
  ]
}
```

生成 SQL：

```sql
AND (product_id = ? AND qty > ? AND discount > 0)
AND (product_id = ? AND qty > ? AND discount > 5)
```

---

### 5️⃣ 动态排序 + val

```sql
SELECT * FROM user
WHERE 1=1
{{if $.params.min_age}}
AND age >= {{val "$.params.min_age"}}
{{end}}
ORDER BY {{val "$.params.sort_field"}} {{val "$.params.sort_order"}}
```

```json
{
  "min_age": 18,
  "sort_field": "created_at",
  "sort_order": "DESC"
}
```

生成 SQL：

```sql
WHERE 1=1 AND age >= 18
ORDER BY created_at DESC
```

---

## 四、能力矩阵

| 功能             | 占位符    | 说明                           |
| -------------- | ------ | ---------------------------- |
| 单字段条件          | expr   | = / in / like / > / < / 空值裁剪 |
| 逻辑组合           | and/or | 任意嵌套，可裁剪空条件                  |
| 条件裁剪           | if     | 控制整段 SQL 是否渲染                |
| 循环生成条件         | for    | 遍历数组生成重复逻辑，可嵌套               |
| 动态字面量 / 内联 SQL | val    | 直接插入值或 SQL 片段，不生成占位符         |

---

## 五、规则总结

1. **空值自动裁剪**：expr / if / for 块中空值不生成 SQL
2. **可嵌套任意层级**：and / or / if / for / expr / val
3. **循环生成**：for 可与 expr / val / and / or 嵌套
4. **val 安全性**：仅用于可信或转义值
5. **BI 查询能力**：用户只需传 JSON，无需写接口逻辑

---

这份 RFC 已经覆盖：

* expr → 单字段条件
* and / or → 逻辑组合
* if → 条件裁剪
* for → 循环生成
* val → 字面量 / 内联值

✅ 可以直接落地到后端 BI 查询引擎，**实现“SQL 即接口 + JSON 参数驱动”**
