package qsql

import "strings"

// cleanSQL 清理 SQL 中的多余空白
func cleanSQL(sql string) string {
	// 分行处理
	lines := strings.Split(sql, "\n")
	var cleaned []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	// 合并并规范化空格
	result := strings.Join(cleaned, " ")

	// 将多个空格替换为单个空格
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return strings.TrimSpace(result)
}

// isEmpty 判断值是否为空
func isEmpty(val interface{}) bool {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []int:
		return len(v) == 0
	case []int64:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	case bool:
		return !v
	default:
		return false
	}
}
