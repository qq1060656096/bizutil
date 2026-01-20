package qsql

// Vars 表示 SQL 执行时使用的变量集合。
// Vars 只负责向 Engine 提供一个 JSON 形式的变量快照，
// 不关心变量的来源、设置方式和内部结构。
type Vars interface {
	// JSON 返回变量的 JSON 字符串表示。
	// 返回值必须是一个合法的 JSON 对象字符串。
	JSON() string
}
