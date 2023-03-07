package TIGER

import "strings"

type InsertBuilder struct {
	//新增语句构建
	builder *strings.Builder
	//新增列
	column []string
	//新增列的值
	args []any
	//新增的表名
	tableName string
}
