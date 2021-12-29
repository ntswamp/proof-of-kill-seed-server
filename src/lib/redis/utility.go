package lib_redis

import (
	lib_error "app/src/lib/error"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// redisの緯度の範囲.
const MinLatitude = -85.05112878
const MaxLatitude = 85.05112878

func marshal(value interface{}) (string, error) {
	s, err := Marshal(value)
	return s, err
}

func unmarshal(redisValue string, dest interface{}) error {
	return Unmarshal(redisValue, dest)
}

func unmarshalSlice(redisValues []string, dest interface{}) error {

	length := len(redisValues)
	if length == 0 {
		return nil
	}

	destV := reflect.ValueOf(dest)
	elem := destV.Elem()
	sl := reflect.MakeSlice(elem.Type(), length, length)

	for i := 0; i < length; i++ {
		item := sl.Index(i)
		if len(redisValues[i]) == 0 {
			if item.Kind() != reflect.Ptr {
				item.Set(reflect.New(item.Type()).Elem())
			}
			continue
		}
		var itemV reflect.Value
		if item.Kind() == reflect.Ptr {
			itemV = reflect.New(item.Type().Elem())
			item.Set(itemV)
		} else {
			item.Set(reflect.New(item.Type()).Elem())
			itemV = item.Addr()
		}
		err := unmarshal(redisValues[i], itemV.Interface())
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	elem.Set(sl)

	return nil
}

func unmarshalMap(redisValues []string, dest interface{}) error {

	length := len(redisValues)
	if length == 0 {
		return nil
	}

	destV := reflect.ValueOf(dest)
	elem := destV.Elem()

	mp := reflect.MakeMap(elem.Type())
	valType := mp.Type().Elem()

	for i := 0; i < len(redisValues); i += 2 {
		key := reflect.ValueOf(redisValues[i])
		redisValue := redisValues[i+1]

		var itemV reflect.Value
		if valType.Kind() == reflect.Ptr {
			itemV = reflect.New(valType.Elem())
			err := unmarshal(redisValue, itemV.Interface())
			if err != nil {
				return lib_error.WrapError(err)
			}
		} else {
			itemV = reflect.New(valType).Elem()
			err := unmarshal(redisValue, itemV.Addr().Interface())
			if err != nil {
				return lib_error.WrapError(err)
			}
		}
		mp.SetMapIndex(key, itemV)
	}
	elem.Set(mp)

	return nil
}

func loadSortedSetMembers(redisStrings []string) []*SortedSetMember {
	dest := []*SortedSetMember{}

	for i := 0; i < len(redisStrings); i += 2 {
		score, _ := strconv.ParseUint(redisStrings[i+1], 10, 64)
		dest = append(dest, &SortedSetMember{
			Name:  redisStrings[i],
			Score: score,
		})
	}

	return dest
}

/**
 * redisの有効な緯度かどうか.
 */
func IsValidLatitute(lat float64) bool {
	return lat <= MaxLatitude && MinLatitude <= lat
}

/**
 * 緯度の丸め込み.
 */
func RoundLatitute(lat float64) float64 {
	return math.Max(math.Min(lat, MaxLatitude), MinLatitude)
}

func Marshal(value interface{}) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	switch string(bytes[:1]) {
	case "[":
		fallthrough
	case "{":
		s := base64.StdEncoding.EncodeToString(bytes)
		return fmt.Sprintf("encoded:%s", s), nil
	default:
		return string(bytes), nil
	}
}

func Unmarshal(src string, dest interface{}) error {
	if len(src) == 0 {
		return nil
	}
	var err error
	var jsonData []byte
	if strings.Index(src, "encoded:") == 0 {
		jsonData, err = base64.StdEncoding.DecodeString(src[8:])
		if err != nil {
			return err
		}
	} else {
		jsonData = []byte(src)
	}
	err = json.Unmarshal(jsonData, dest)
	return err
}
