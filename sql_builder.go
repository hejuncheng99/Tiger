package TIGER

import (
	"database/sql"
	"strings"
)

type SqlBuilder struct {
	// 查询语句构建
	Builder *strings.Builder
	//需要查询的列
	column []string
	//查询的表名
	tableName string
	//查询条件
	//where []func(s *SqlBuilder)
	where *WhereExpr
	//查询参数
	args []any
	//排序语句
	orderBy string
	//偏移量
	offSet *int64
	//限制查询的数据数量
	limit *int64
	//分组语句
	groupBy string
	//聚合
	having string
	//受影响行数
	RowsAffected int64
	//查询返回行
	rows *sql.Rows
}

type WhereExpr struct {
	SQL  string
	Vars []any
}
