package TIGER

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Config struct {
}
type DB struct {
	*Config
	Db       *sql.DB
	sqlBuild *SqlBuilder
}

// TigerEngine 引擎
type TigerEngine struct {
}

func NewMysql(Username string, Password string, Address string, Dbname string) (*DB, error) {
	dsn := Username + ":" + Password + "@tcp(" + Address + ")/" + Dbname + "?charset=utf8&parseTime=True&timeout=5s&readTimeout=6s"
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	//最大连接数等配置，先占个位
	//2.0版本支持了其他数据库呢；
	//后续连接池的加入

	//db.SetMaxOpenConns(3)
	//db.SetMaxIdleConns(3)

	return &DB{
		Db:       db,
		sqlBuild: &SqlBuilder{Builder: &strings.Builder{}},
	}, nil
}

//func (e *TigerEngine) Save(v any) (ex *TigerEngine) {
//
//	db := TigerEngine{}
//	db.Transaction(func(tx *sql.Tx) error {
//		tx.
//	})
//}

// Transaction 事务操作
func (e *DB) Transaction(fx func(tx *sql.Tx) error) error {
	tx, err := e.Db.Begin()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	isCommit := true

	//实际操作
	if err := fx(tx); err != nil {
		isCommit = false
		return err
	}

	if isCommit {
		err = tx.Commit()
		fmt.Println("transaction commit")
		return err
	} else {
		err = tx.Rollback()
		fmt.Println("transaction rollback")
		return err
	}
}

// Select 选取字段构建
func (tg *DB) Select(filed ...string) *DB {
	tg.sqlBuild.column = append(tg.sqlBuild.column, filed...)
	return tg
}

// From 查询表构建
func (tg *DB) From(name string) *DB {
	tg.sqlBuild.tableName = name
	return tg
}

// Where 查询字段构建
func (tg *DB) Where(filed string, args ...any) *DB {
	need := strings.Count(filed, "?")
	got := len(args)
	if need != got {
		log.Fatalf("statements expected %v arguments, got %v", need, got)
	}
	var w = new(WhereExpr)
	w.SQL = filed
	w.Vars = args
	tg.sqlBuild.where = w
	return tg
}

// OrderBy 排序字段构建
func (tg *DB) OrderBy(field string) *DB {
	tg.sqlBuild.orderBy = field
	return tg
}

// OffSet 偏移量构建
func (tg *DB) OffSet(offset int64) *DB {
	tg.sqlBuild.offSet = &offset
	return tg
}

// Limit 限制查询数据的条数
func (tg *DB) Limit(limit int64) *DB {
	tg.sqlBuild.limit = &limit
	return tg
}

// Query 查询语句构建
func (tg *DB) Query() *DB {
	tg.sqlBuild.Builder.WriteString("SELECT ")
	for i, v := range tg.sqlBuild.column {
		if i > 0 {
			tg.sqlBuild.Builder.WriteString(",")
		}
		tg.sqlBuild.Builder.WriteString(v)
	}
	tg.sqlBuild.Builder.WriteString(" FROM ")

	tg.sqlBuild.Builder.WriteString(tg.sqlBuild.tableName + " ")

	if len(tg.sqlBuild.where.SQL) > 0 {
		tg.sqlBuild.Builder.WriteString(" WHERE ")

		tg.sqlBuild.Builder.WriteString(tg.sqlBuild.where.SQL)
		tg.sqlBuild.args = tg.sqlBuild.where.Vars
	}

	if tg.sqlBuild.orderBy != "" {
		tg.sqlBuild.Builder.WriteString(" ORDER BY " + tg.sqlBuild.orderBy)
	}

	if tg.sqlBuild.limit != nil {
		tg.sqlBuild.Builder.WriteString(" LIMIT ")
		tg.sqlBuild.Builder.WriteString(strconv.FormatInt(*tg.sqlBuild.limit, 10))
	}

	if tg.sqlBuild.offSet != nil {
		tg.sqlBuild.Builder.WriteString(" OFFSET ")
		tg.sqlBuild.Builder.WriteString(strconv.FormatInt(*tg.sqlBuild.limit, 10))
	}

	log.Printf(tg.sqlBuild.Builder.String())
	rows, err := tg.Db.Query(tg.sqlBuild.Builder.String(), tg.sqlBuild.args...)
	if err != nil {
		log.Println(err)
	}
	tg.sqlBuild.rows = rows

	return tg
}

// ScanRows 查询结果转换
func (tg *DB) ScanRows(dst any) error {
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("scan rows close error : %v", err.Error())
		}
	}(tg.sqlBuild.rows)

	//dts 获取需要转换的slice的指针地址
	val := reflect.ValueOf(dst)

	//判断是否是指针类型，go是指传递，只有传指针才能让更改生效
	if val.Kind() != reflect.Ptr {
		return DtsNotPointerError
	}

	//指针指向value 获取具体的值
	val = reflect.Indirect(val)
	if val.Kind() != reflect.Slice {
		return DtsNotSlice
	}

	//获取slice中的类型
	strPointer := val.Type().Elem()

	//指针指向的类型，具体结构体
	str := strPointer.Elem()

	columns, err := tg.sqlBuild.rows.Columns()
	if err != nil {
		return err
	}

	//结构体的json tag 的value 对应字段在结构体中的index
	// map tag -》 field idx
	tagIdx := make(map[string]int)

	//结构体的torm tag 的value对应结构体中的index
	for i := 0; i < str.NumField(); i++ {
		tagName := str.Field(i).Tag.Get("torm")
		if tagName != "" {
			tagIdx[tagName] = i
		}
	}

	//字段类型
	resultType := make([]reflect.Type, 0, len(columns))
	//字段在结构体中的序号
	index := make([]int, 0, len(columns))

	for _, v := range columns {
		if i, ok := tagIdx[v]; ok {
			resultType = append(resultType, str.Field(i).Type)
			index = append(index, i)
		}
	}

	for tg.sqlBuild.rows.Next() {
		//创建结构体指针，获取指针指向的对象
		obj := reflect.New(str).Elem()
		result := make([]any, 0, len(resultType))

		//创建结构体字段类型实例的指针，并转化为interface{}类型
		for _, v := range resultType {
			result = append(result, reflect.New(v).Interface())
		}

		//扫描结果
		err := tg.sqlBuild.rows.Scan(result...)
		if err != nil {
			return err
		}

		for i, v := range result {
			//找到对应的结构体index
			fieldIndex := index[i]
			//把scan后的值通过反射得到指针指向的value，赋值给对应的结构体
			obj.Field(fieldIndex).Set(reflect.ValueOf(v).Elem())
		}

		//append到slice中
		vv := reflect.Append(val, obj.Addr())
		val.Set(vv)
	}

	return tg.sqlBuild.rows.Err()
}

func (tg *DB) Insert(obj any) int64 {
	tg.sqlBuild.Builder.WriteString("INSERT INTO")

	tg.sqlBuild.Builder.WriteString(" " + tg.sqlBuild.tableName + "(")

	for i, col := range tg.sqlBuild.column {
		if i > 0 {
			tg.sqlBuild.Builder.WriteString(",")
		}
		tg.sqlBuild.Builder.WriteString(col)
	}

	tg.sqlBuild.Builder.WriteString(")  ")

	tg.sqlBuild.Builder.WriteString("VALUES")

	//for _, arg := range tg.sqlBuild.args {
	//	if  {
	//
	//	}
	//}

	return 1
}
