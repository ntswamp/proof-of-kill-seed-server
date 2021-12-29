package lib_db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

type ModelInfo struct {
	TableName          string
	PrimaryFields      []*gorm.StructField
	StructFields       map[string]*gorm.StructField
	customTypes        map[string]CustomType
	customSelectClause string
	cacheStore         *ModelCacheStore
	dbNameMap          map[string]string
}

func NewModelInfo(db *gorm.DB, model interface{}) *ModelInfo {
	scope := db.NewScope(model)
	modelStruct := scope.GetModelStruct()
	info := &ModelInfo{
		TableName:     scope.TableName(),
		PrimaryFields: modelStruct.PrimaryFields,
		StructFields:  make(map[string]*gorm.StructField, len(modelStruct.StructFields)),
		customTypes:   map[string]CustomType{},
		cacheStore:    nil,
		dbNameMap:     make(map[string]string, len(modelStruct.StructFields)),
	}
	for _, field := range modelStruct.StructFields {
		info.StructFields[field.Name] = field
		// customTypes.
		indirectType := field.Struct.Type
		for indirectType.Kind() == reflect.Ptr {
			indirectType = indirectType.Elem()
		}
		fieldValue := reflect.New(indirectType).Interface()
		switch fieldValue.(type) {
		case *Point:
			info.customTypes[field.Name] = fieldValue.(*Point)
		}
		info.dbNameMap[field.DBName] = field.Name
	}
	return info
}

/**
 * プライマリキーの長さ.
 */
func (self *ModelInfo) KeyLen() int {
	return len(self.PrimaryFields)
}

/**
 * このmodelのModelCacheStoreのインスタンスを取得.
 */
func (self *ModelInfo) CacheStore() *ModelCacheStore {
	if self.cacheStore == nil {
		self.cacheStore = NewModelCacheStore(self)
	}
	return self.cacheStore
}

/**
 * プライマリキーを文字列化.
 */
func (self *ModelInfo) MakePrimaryKeyString(pkey interface{}) string {
	var keys []string
	switch pkey.(type) {
	case []string:
		keys = pkey.([]string)
	default:
		keyValue := reflect.ValueOf(pkey)
		switch keyValue.Kind() {
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			keys = make([]string, keyValue.Len())
			for i := 0; i < keyValue.Len(); i++ {
				keys[i] = fmt.Sprintf("%v", keyValue.Index(i).Interface())
			}
		default:
			keys = []string{fmt.Sprintf("%v", pkey)}
		}
	}
	return strings.Join(keys, "#&#")
}

/**
 * プライマリキーを文字列化.
 */
func (self *ModelInfo) MakePrimaryKeyStringByModel(model interface{}, length int) string {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = reflect.Indirect(modelValue)
	}
	keyLen := self.KeyLen()
	if length < 1 || keyLen < length {
		length = keyLen
	}
	pkeyStrings := make([]string, length)
	for i := 0; i < length; i++ {
		field := self.PrimaryFields[i]
		pkeyStrings[i] = fmt.Sprintf("%v", modelValue.FieldByName(field.Name).Interface())
	}
	return self.MakePrimaryKeyString(pkeyStrings)
}

/**
 * マスターデータかどうかを取得.
 */
func (self *ModelInfo) GetIsMaster(model interface{}) bool {
	modelValue := reflect.Indirect(reflect.ValueOf(model))
	if modelValue.Kind() == reflect.Array || modelValue.Kind() == reflect.Slice {
		tmp := reflect.Indirect(reflect.New(modelValue.Type().Elem()))
		if tmp.Kind() == reflect.Ptr {
			tmp = reflect.Indirect(reflect.New(tmp.Type().Elem()))
		}
		modelValue = tmp
	}
	getIsMaster := modelValue.MethodByName("GetIsMaster")
	if !getIsMaster.IsValid() {
		return false
	}
	return getIsMaster.Call([]reflect.Value{})[0].Interface().(bool)
}

/**
 * CustomeTypeの取得.
 */
func (self *ModelInfo) HasCustomTypes() bool {
	return 0 < len(self.customTypes)
}

/**
 * CustomeTypeの取得.
 */
func (self *ModelInfo) GetCustomType(name string) CustomType {
	customType, _ := self.customTypes[name]
	if customType != nil {
		return customType
	}
	name, _ = self.dbNameMap[name]
	customType, _ = self.customTypes[name]
	return customType
}

/**
 * SQLのSelect句を取得.
 */
func (self *ModelInfo) getCustomSelectClause() string {
	if 0 < len(self.customSelectClause) {
		return self.customSelectClause
	}
	columns := make([]string, len(self.StructFields))
	i := 0
	for fieldName, field := range self.StructFields {
		customType, _ := self.customTypes[fieldName]
		if customType != nil {
			columns[i] = customType.Select(field.DBName)
		} else {
			columns[i] = fmt.Sprintf("`%s`", field.DBName)
		}
		i++
	}
	self.customSelectClause = strings.Join(columns, ",")
	return self.customSelectClause
}

/**
 * SELECT句を付ける.
 */
func (self *ModelInfo) Select(db *gorm.DB) *gorm.DB {
	if 0 < len(self.customTypes) {
		db = db.Select(self.getCustomSelectClause())
	}
	return db
}

/**
 * プライマリキーで検索するクエリを付ける.
 */
func (self *ModelInfo) Query(db *gorm.DB, keys []interface{}, keyLength int) *gorm.DB {
	db = db.Table(self.TableName)
	if self.KeyLen() < keyLength {
		keyLength = self.KeyLen()
	}
	if keyLength == 1 {
		// キーの長さが1の場合はシンプル.
		if len(keys) == 1 {
			return db.Where(fmt.Sprintf("\"%s\" = ?", self.PrimaryFields[0].DBName), keys[0])
		} else {
			return db.Where(fmt.Sprintf("\"%s\" in (?)", self.PrimaryFields[0].DBName), keys)
		}
	}
	sarr := make([]string, keyLength)
	for i, _ := range sarr {
		sarr[i] = self.PrimaryFields[i].DBName
	}
	varr := make([]string, len(keys))
	for i, _ := range varr {
		varr[i] = "(?)"
	}
	query := fmt.Sprintf("(`%s`) in (%s)", strings.Join(sarr, "`,`"), strings.Join(varr, "`,`"))
	return db.Where(query, keys...)
}

/**
 * モデルのプライマリキーで検索するクエリを付ける.
 */
func (self *ModelInfo) QueryByModel(db *gorm.DB, models interface{}) *gorm.DB {
	modelsValue := reflect.ValueOf(models)
	if modelsValue.Kind() != reflect.Array && modelsValue.Kind() != reflect.Slice {
		tmp := reflect.New(reflect.SliceOf(modelsValue.Type())).Elem()
		tmp = reflect.Append(tmp, modelsValue)
		modelsValue = tmp
	}
	keys := make([]interface{}, modelsValue.Len())
	for i := 0; i < modelsValue.Len(); i++ {
		modelValue := modelsValue.Index(i)
		if modelValue.Kind() == reflect.Ptr {
			modelValue = reflect.Indirect(modelValue)
		}
		pkeys := make([]interface{}, self.KeyLen())
		for j, field := range self.PrimaryFields {
			pkeys[j] = modelValue.FieldByName(field.Name).Interface()
		}
		keys[i] = pkeys
	}
	return self.Query(db, keys, self.KeyLen()).Model(models)
}
