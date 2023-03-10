package TIGER

import (
	"database/sql"
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Config struct {
}
type DB struct {
	*Config
	sqlBuild *SqlBuilder
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
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)

	return &DB{
		sqlBuild: &SqlBuilder{Builder: &strings.Builder{}, Db: db},
	}, nil
}

// Transaction 事务操作
func (tg *DB) Transaction(fx func(db *DB) error) error {
	tx, err := tg.sqlBuild.Db.Begin()
	tg.sqlBuild.Tx = tx
	if err != nil {
		_ = tx.Rollback()
		log.Printf("open transaction fail : %v", err)
		return err
	}
	//实际操作
	if err := fx(tg); err != nil {
		_ = tx.Rollback()
		log.Printf("run transaction fail : %v", err)
		return err
	}
	err = tx.Commit()
	return nil
}

// Select 选取字段构建
func (tg *DB) Select(filed ...string) *DB {
	tg.sqlBuild.column = append(tg.sqlBuild.column, filed...)
	return tg
}

// From 查询表构建
func (tg *DB) Table(name string) *DB {
	if tg.sqlBuild.Builder.Len() != 0 {
		if tg.sqlBuild.Tx != nil {
			tg = &DB{
				sqlBuild: &SqlBuilder{
					Db:      tg.sqlBuild.Db,
					Builder: &strings.Builder{},
					Tx:      tg.sqlBuild.Tx,
				},
			}
		} else {
			tg = &DB{
				sqlBuild: &SqlBuilder{
					Db:      tg.sqlBuild.Db,
					Builder: &strings.Builder{},
				},
			}
		}

	}

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

	//log.Printf(tg.sqlBuild.Builder.String())
	rows, err := tg.sqlBuild.Db.Query(tg.sqlBuild.Builder.String(), tg.sqlBuild.args...)
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
	if val.IsNil() {
		return errors.New("参数不能是空指针！")
	}

	//指针指向value 获取具体的值
	val = reflect.Indirect(val)
	if val.Kind() != reflect.Slice {
		return DtsNotSlice
	}

	//原始单个struct的类型
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
		tagName := str.Field(i).Tag.Get("sql")
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

// Insert 插入数据
func (tg *DB) Insert(obj any) error {
	tg.sqlBuild.Builder.WriteString("INSERT INTO")
	tg.sqlBuild.Builder.WriteString(" " + tg.sqlBuild.tableName)
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		//切片大小
		l := v.Len()
		for i := 0; i < l; i++ {
			v := v.Index(i)
			t := v.Type()
			if i > 0 {
				tg.sqlBuild.Builder.WriteString(",")
			}
			var placeholder []string
			var fields []string
			for j := 0; j < v.NumField(); j++ {
				//获取字段
				if i == 0 {
					fields = append(fields, t.Field(j).Tag.Get("sql"))
				}
				//数据占位符和数据写入
				placeholder = append(placeholder, "?")
				tg.sqlBuild.args = append(tg.sqlBuild.args, v.Field(j).Interface())

			}
			if i == 0 {
				//字段写入
				tg.sqlBuild.Builder.WriteString("(")
				tg.sqlBuild.Builder.WriteString(strings.Join(fields, ","))
				tg.sqlBuild.Builder.WriteString(")")
				tg.sqlBuild.Builder.WriteString(" VALUES ")
			}
			//占位符
			tg.sqlBuild.Builder.WriteString("(")
			tg.sqlBuild.Builder.WriteString(strings.Join(placeholder, ","))
			tg.sqlBuild.Builder.WriteString(")")
		}
	case reflect.Struct:
		//字段列表
		var placeholder []string
		var fields []string
		for i := 0; i < t.NumField(); i++ {
			fields = append(fields, t.Field(i).Tag.Get("sql"))
			placeholder = append(placeholder, "?")
			tg.sqlBuild.args = append(tg.sqlBuild.args, v.Field(i).Interface())
		}

		tg.sqlBuild.Builder.WriteString("(")
		tg.sqlBuild.Builder.WriteString(strings.Join(fields, ","))
		tg.sqlBuild.Builder.WriteString(")")
		tg.sqlBuild.Builder.WriteString("VALUE")
		tg.sqlBuild.Builder.WriteString("(")
		tg.sqlBuild.Builder.WriteString(strings.Join(placeholder, ","))
		tg.sqlBuild.Builder.WriteString(")")
	}

	exp := tg.sqlBuild.Builder.String()
	//log.Printf("SQL:%v", exp)
	stmt, err := tg.sqlBuild.Db.Prepare(exp)
	if err != nil {
		log.Printf("Prepare Error :%v", err)
		return err
	}
	exec, err := stmt.Exec(tg.sqlBuild.args...)
	if err != nil {
		log.Printf("Prepare Error :%v", err)
		return err
	}
	id, _ := exec.LastInsertId()
	log.Printf("exec.LastInsertId()：%v", id)
	affected, _ := exec.RowsAffected()
	log.Printf("affected：%v", affected)

	return nil
}

// Update 更新
func (tg *DB) Update(obj ...any) (int64, error) {
	var dataType int
	if len(obj) == 1 {
		dataType = 1
	} else if len(obj) == 2 {
		dataType = 2
	} else {
		return 0, errors.New("参数个数错误")
	}

	tg.sqlBuild.Builder.WriteString("UPDATE  ")
	tg.sqlBuild.Builder.WriteString(tg.sqlBuild.tableName)
	tg.sqlBuild.Builder.WriteString(" SET ")

	//如果是结构体
	if dataType == 1 {
		t := reflect.TypeOf(obj[0])
		v := reflect.ValueOf(obj[0])
		var fieldNameArray []string
		for i := 0; i < t.NumField(); i++ {
			//首字母小写，不可反射
			if !v.Field(i).CanInterface() {
				continue
			}
			//解析tag，找出真实的sql字段名
			sqlTag := t.Field(i).Tag.Get("sql")
			if sqlTag != "" {
				fieldNameArray = append(fieldNameArray, strings.Split(sqlTag, ",")[0]+"=?")
			} else {
				fieldNameArray = append(fieldNameArray, t.Field(i).Name+"=?")
			}

			tg.sqlBuild.args = append(tg.sqlBuild.args, v.Field(i).Interface())
		}
		tg.sqlBuild.Builder.WriteString(strings.Join(fieldNameArray, ","))

	} else if dataType == 2 {
		//直接=的情况
		tg.sqlBuild.Builder.WriteString(obj[0].(string) + "=?")
		tg.sqlBuild.args = append(tg.sqlBuild.args, obj[1])
	}

	if tg.sqlBuild.where.SQL != "" {
		tg.sqlBuild.Builder.WriteString(" WHERE ")
		tg.sqlBuild.Builder.WriteString(tg.sqlBuild.where.SQL)
		tg.sqlBuild.args = append(tg.sqlBuild.args, tg.sqlBuild.where.Vars...)
	}

	if tg.sqlBuild.limit != nil {
		tg.sqlBuild.Builder.WriteString(" LIMIT " + strconv.FormatInt(*tg.sqlBuild.limit, 10))
	}
	expr := tg.sqlBuild.Builder.String()
	//log.Printf("SQL:%v", expr)
	stmt, err := tg.sqlBuild.Db.Prepare(expr)
	if err != nil {
		log.Printf("update prepare err:%v", err.Error())
		return 0, err
	}

	result, err := stmt.Exec(tg.sqlBuild.args...)
	if err != nil {
		log.Printf("update exec error : %v", err)
		return 0, err
	}
	affected, _ := result.RowsAffected()

	return affected, nil
}

// Delete 删除记录
func (tg *DB) Delete() (int64, error) {
	tg.sqlBuild.Builder.WriteString("DELETE FROM ")
	tg.sqlBuild.Builder.WriteString(tg.sqlBuild.tableName)

	if tg.sqlBuild.where.SQL != "" {
		tg.sqlBuild.Builder.WriteString(" WHERE " + tg.sqlBuild.where.SQL)
	}

	if tg.sqlBuild.limit != nil {
		tg.sqlBuild.Builder.WriteString(" LIMIT " + strconv.FormatInt(*tg.sqlBuild.limit, 10))
	}
	expr := tg.sqlBuild.Builder.String()
	//log.Printf("SQL:%v", expr)
	stmt, err := tg.sqlBuild.Db.Prepare(expr)
	if err != nil {
		log.Printf("delete prepare error:%v", err)
		return 0, err
	}

	result, err := stmt.Exec(tg.sqlBuild.where.Vars...)
	if err != nil {
		log.Printf("delete exec error:%v", err)
		return 0, err
	}
	//影响的行数
	rowsAffected, _ := result.RowsAffected()

	return rowsAffected, nil
}
