package TIGER

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// InterfaceToString interface 2 string util
func InterfaceToString(v any) string {
	var key string
	switch v.(type) {
	case float64:
		ft := v.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := v.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := v.(int)
		key = strconv.Itoa(it)
	case uint:
		it := v.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := v.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := v.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := v.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := v.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := v.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := v.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := v.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := v.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = v.(string)
	case time.Time:
		t, _ := v.(time.Time)
		key = t.String()
		// 2022-11-23 11:29:07 +0800 CST  这类格式把尾巴去掉
		key = strings.Replace(key, " +0800 CST", "", 1)
	case []byte:
		key = string(v.([]byte))
	default:
		newValue, _ := json.Marshal(v)
		key = string(newValue)
	}

	return key

}
