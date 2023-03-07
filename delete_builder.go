package TIGER

import "strings"

type DeleteBuilder struct {
	//删除语句构建
	builder *strings.Builder
	//需要删除的表名
	tableName string
	//查询条件
	where []func(d *DeleteBuilder)
	//查询参数
	args []any
	//受影响行数
	RowsAffected int64
}
