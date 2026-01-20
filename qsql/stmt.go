package qsql

// SQLStmt 执行结果
type SQLStmt struct {
	RawSQL           string
	SQL              string            // 生成的 SQL
	Args             []interface{}     // 参数列表
	Errors           []string          // 错误列表（记录缺失的参数等）
	ValidatorsErrors []*ValidatorError // 验证器错误列表
}

func (s *SQLStmt) HasValidatorErrors() bool {
	return len(s.ValidatorsErrors) > 0
}

func (s *SQLStmt) HasErrors() bool {
	return len(s.Errors) > 0
}
