package TIGER

//import (
//	"database/sql"
//	"log"
//	"reflect"
//)
//
//func ScanRows(rows *sql.Rows, dst any) error {
//	defer func(rows *sql.Rows) {
//		err := rows.Close()
//		if err != nil {
//			log.Printf("scan rows close error : %v", err.Error())
//		}
//	}(rows)
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
//	columns, err := rows.Columns()
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
//	for rows.Next() {
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
//		err := rows.Scan(result...)
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
//	return rows.Err()
//}
