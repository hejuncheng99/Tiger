package TIGER

import (
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// SelectBuilder query builder
type SelectBuilder struct {
	// 查询语句构建
	builder *strings.Builder
	//需要查询的列
	column []string
	//查询的表名
	tableName string
	//查询条件
	where []func(s *SelectBuilder)
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
}

// Select 选取字段构建
func (s *SelectBuilder) Select(filed ...string) *SelectBuilder {
	s.column = append(s.column, filed...)
	return s
}

// From 查询表构建
func (s *SelectBuilder) From(name string) *SelectBuilder {
	s.tableName = name
	return s
}

// Where 查询字段构建
func (s *SelectBuilder) Where(f ...func(s *SelectBuilder)) *SelectBuilder {
	s.where = append(s.where, f...)
	return s
}

// OrderBy 排序字段构建
func (s *SelectBuilder) OrderBy(field string) *SelectBuilder {
	s.orderBy = field
	return s
}

// OffSet 偏移量构建
func (s *SelectBuilder) OffSet(offset int64) *SelectBuilder {
	s.offSet = &offset
	return s
}

// Limit 限制查询数据的条数
func (s *SelectBuilder) Limit(limit int64) *SelectBuilder {
	s.limit = &limit
	return s
}

// GT 大于
func GT(field string, arg any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " > ?")
		s.args = append(s.args, arg)
	}
}

// Eq 等于
func Eq(field string, arg any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " = ?")
		s.args = append(s.args, arg)
	}
}

// Ne 不等于
func Ne(field string, arg any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " <> ?")
		s.args = append(s.args, arg)
	}
}

// Ge 大于等于
func Ge(field string, arg any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " >= ?")
		s.args = append(s.args, arg)
	}
}

// Lt 小于
func Lt(field string, arg any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " < ?")
		s.args = append(s.args, arg)
	}
}

// Le 小于等于
func Le(field string, arg any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " <= ?")
		s.args = append(s.args, arg)
	}
}

// Between
func Between(field string, arg1 any, arg2 any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " between ? and ?")
		s.args = append(s.args, arg1)
		s.args = append(s.args, arg2)
	}
}

func Like(field string, arg any) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " LIKE %?%")
		s.args = append(s.args, arg)
	}
}

// Query 查询语句构建
func (s *SelectBuilder) Query() (string, []any) {
	s.builder.WriteString("SELECT ")
	for i, v := range s.column {
		if i > 0 {
			s.builder.WriteString(",")
		}
		s.builder.WriteString("`" + v + "`")
	}
	s.builder.WriteString(" FROM ")

	s.builder.WriteString("`" + s.tableName + "`")

	if len(s.where) > 0 {
		s.builder.WriteString(" WHERE ")
		for i, v := range s.where {
			fn := runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
			if i > 0 && fn == "Or" {
				s.builder.WriteString(" OR ")
			} else {
				if i > 0 {
					s.builder.WriteString(" AND ")
				}
			}

			v(s)
		}
	}

	if s.orderBy != "" {
		s.builder.WriteString(" ORDER BY " + s.orderBy)
	}

	if s.limit != nil {
		s.builder.WriteString(" LIMIT ")
		s.builder.WriteString(strconv.FormatInt(*s.limit, 10))
	}

	if s.offSet != nil {
		s.builder.WriteString(" OFFSET ")
		s.builder.WriteString(strconv.FormatInt(*s.limit, 10))
	}

	return s.builder.String(), s.args
}
