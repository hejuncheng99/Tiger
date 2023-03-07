package TIGER

import (
	"database/sql"
	"strings"
)

// SelectBuilder query builder
type SelectBuilder struct {
	// 查询语句构建
	Builder *strings.Builder
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
	//聚合
	having string
	//查询返回行
	rows *sql.Rows
}

//
//// Select 选取字段构建
//func (s *SelectBuilder) Select(filed ...string) *SelectBuilder {
//	s.column = append(s.column, filed...)
//	return s
//}
//
//// From 查询表构建
//func (s *SelectBuilder) From(name string) *SelectBuilder {
//	s.tableName = name
//	return s
//}
//
//// Where 查询字段构建
//func (s *SelectBuilder) Where(f ...func(s *SelectBuilder)) *SelectBuilder {
//	s.where = append(s.where, f...)
//	return s
//}
//
//// OrderBy 排序字段构建
//func (s *SelectBuilder) OrderBy(field string) *SelectBuilder {
//	s.orderBy = field
//	return s
//}
//
//// OffSet 偏移量构建
//func (s *SelectBuilder) OffSet(offset int64) *SelectBuilder {
//	s.offSet = &offset
//	return s
//}
//
//// Limit 限制查询数据的条数
//func (s *SelectBuilder) Limit(limit int64) *SelectBuilder {
//	s.limit = &limit
//	return s
//}
//
//// GT 大于
//func GT(field string, arg any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " > ?")
//		s.args = append(s.args, arg)
//	}
//}
//
//// Eq 等于
//func Eq(field string, arg any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " = ?")
//		s.args = append(s.args, arg)
//	}
//}
//
//// Ne 不等于
//func Ne(field string, arg any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " <> ?")
//		s.args = append(s.args, arg)
//	}
//}
//
//// Ge 大于等于
//func Ge(field string, arg any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " >= ?")
//		s.args = append(s.args, arg)
//	}
//}
//
//// Lt 小于
//func Lt(field string, arg any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " < ?")
//		s.args = append(s.args, arg)
//	}
//}
//
//// Le 小于等于
//func Le(field string, arg any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " <= ?")
//		s.args = append(s.args, arg)
//	}
//}
//
//// Between
//func Between(field string, arg1 any, arg2 any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " between ? and ?")
//		s.args = append(s.args, arg1)
//		s.args = append(s.args, arg2)
//	}
//}
//
//func Like(field string, arg any) func(s *SelectBuilder) {
//	return func(s *SelectBuilder) {
//		s.Builder.WriteString(field + " LIKE ?")
//		s.args = append(s.args, arg)
//	}
//}
//
//// Query 查询语句构建
//func (s *SelectBuilder) Query() *SelectBuilder {
//	s.Builder.WriteString("SELECT ")
//	for i, v := range s.column {
//		if i > 0 {
//			s.Builder.WriteString(",")
//		}
//		s.Builder.WriteString(v)
//	}
//	s.Builder.WriteString(" FROM ")
//
//	s.Builder.WriteString(s.tableName + " ")
//
//	if len(s.where) > 0 {
//		s.Builder.WriteString(" WHERE ")
//		for i, v := range s.where {
//			fn := runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
//			if i > 0 && fn == "Or" {
//				s.Builder.WriteString(" OR ")
//			} else {
//				if i > 0 {
//					s.Builder.WriteString(" AND ")
//				}
//			}
//
//			v(s)
//		}
//	}
//
//	if s.orderBy != "" {
//		s.Builder.WriteString(" ORDER BY " + s.orderBy)
//	}
//
//	if s.limit != nil {
//		s.Builder.WriteString(" LIMIT ")
//		s.Builder.WriteString(strconv.FormatInt(*s.limit, 10))
//	}
//
//	if s.offSet != nil {
//		s.Builder.WriteString(" OFFSET ")
//		s.Builder.WriteString(strconv.FormatInt(*s.limit, 10))
//	}
//
//	log.Printf(s.Builder.String())
//
//	//  s.Builder.String(), s.args
//
//	return s
//}
//
//// 查询结果转换
//func (s *SelectBuilder) ScanRows(dst any) error {
//	defer func(rows *sql.Rows) {
//		err := rows.Close()
//		if err != nil {
//			log.Printf("scan rows close error : %v", err.Error())
//		}
//	}(s.rows)
//
//	//dts 获取需要转换的slice的指针地址
//	val := reflect.ValueOf(dst)
//
//	//判断是否是指针类型，go是指传递，只有传指针才能让更改生效
//	if val.Kind() != reflect.Ptr {
//		return DtsNotPointerError
//	}
//
//	//指针指向value 获取具体的值
//	val = reflect.Indirect(val)
//	if val.Kind() != reflect.Slice {
//		return DtsNotSlice
//	}
//
//	//获取slice中的类型
//	strPointer := val.Type().Elem()
//
//	//指针指向的类型，具体结构体
//	str := strPointer.Elem()
//
//	columns, err := s.rows.Columns()
//	if err != nil {
//		return err
//	}
//
//	//结构体的json tag 的value 对应字段在结构体中的index
//	// map tag -》 field idx
//	tagIdx := make(map[string]int)
//
//	//结构体的json tag 的value对应结构体中的index
//	for i := 0; i < str.NumField(); i++ {
//		tagName := str.Field(i).Tag.Get("torm")
//		if tagName != "" {
//			tagIdx[tagName] = i
//		}
//	}
//
//	//字段类型
//	resultType := make([]reflect.Type, 0, len(columns))
//	//字段在结构体中的序号
//	index := make([]int, 0, len(columns))
//
//	for _, v := range columns {
//		if i, ok := tagIdx[v]; ok {
//			resultType = append(resultType, str.Field(i).Type)
//			index = append(index, i)
//		}
//	}
//
//	for s.rows.Next() {
//		//创建结构体指针，获取指针指向的对象
//		obj := reflect.New(str).Elem()
//		result := make([]any, 0, len(resultType))
//
//		//创建结构体字段类型实例的指针，并转化为interface{}类型
//		for _, v := range resultType {
//			result = append(result, reflect.New(v).Interface())
//		}
//
//		//扫描结果
//		err := s.rows.Scan(result...)
//		if err != nil {
//			return err
//		}
//
//		for i, v := range result {
//			//找到对应的结构体index
//			fieldIndex := index[i]
//			//把scan后的值通过反射得到指针指向的value，赋值给对应的结构体
//			obj.Field(fieldIndex).Set(reflect.ValueOf(v).Elem())
//		}
//
//		//append到slice中
//		vv := reflect.Append(val, obj.Addr())
//		val.Set(vv)
//	}
//
//	return s.rows.Err()
//}
