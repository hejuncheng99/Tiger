package TIGER

import "strings"

type UpdateBuilder struct {
	//更新语句构建
	builder *strings.Builder
	//需要更新的列
	column []string
	//需要更新的表名
	tableName string
	//查询条件
	where []func(u *UpdateBuilder)
	//查询参数
	args []any
	//受影响行数
	RowsAffected int64
}
