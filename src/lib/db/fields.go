/*
	Custom Fields.
*/
package lib_db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type CustomType interface {
	Select(columnName string) string          // Select時のカラム指定を書き換える用.
	Update(value interface{}) (string, error) // Insert,Update時のカラム指定を書き換える用.
}

/**
 * MySQLのPoint型を使うための型.
 */
type Point struct {
	Lat float64
	Lon float64
}

func (self *Point) Select(columnName string) string {
	return fmt.Sprintf("ST_AsText(`%s`) as `%s`", columnName, columnName)
}

func (self *Point) Update(value interface{}) (string, error) {
	var v *Point
	switch value.(type) {
	case *Point:
		v = value.(*Point)
	case Point:
		tmp := value.(Point)
		v = &tmp
	default:
		return "", errors.New("Point型ではありません")
	}
	return fmt.Sprintf("ST_GeomFromText('POINT(%v %v)')", v.Lat, v.Lon), nil
}

func (self Point) Value() (driver.Value, error) {
	return driver.Value(self), errors.New("Pointを保存するときはModelManagerを使用してください")
}
func (self *Point) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	// ST_AsTextで取得した値になる想定.POINT(lat lon).
	s := string(value.([]uint8))
	s = strings.Replace(s, "POINT(", "", 1)
	s = strings.Replace(s, ")", "", 1)
	data := strings.Split(s, " ")
	if len(data) != 2 {
		// 未設定扱い.
		return nil
	}
	var err error = nil
	self.Lat, err = strconv.ParseFloat(data[0], 64)
	if err != nil {
		return err
	}
	self.Lon, err = strconv.ParseFloat(data[1], 64)
	if err != nil {
		return err
	}
	return nil
}
