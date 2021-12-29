package lib_db

import (
	"github.com/jinzhu/gorm"
)

var modelInfos = map[string]*ModelInfo{}

// gorm.ScopeのchangeableFieldと同じ.
func changeableField(scope *gorm.Scope, field *gorm.Field) bool {
	if selectAttrs := scope.SelectAttrs(); len(selectAttrs) > 0 {
		for _, attr := range selectAttrs {
			if field.Name == attr || field.DBName == attr {
				return true
			}
		}
		return false
	}
	for _, attr := range scope.OmitAttrs() {
		if field.Name == attr || field.DBName == attr {
			return false
		}
	}
	return true
}

// gorm/utilのaddExtraSpaceIfExistと同じ.
func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

// modelinfoの取得.
// callbackで使うmodelInfoをglobalに保持しておく.
func getModelInfo(db *gorm.DB, model interface{}) *ModelInfo {
	name := db.NewScope(model).QuotedTableName()
	if _, ok := modelInfos[name]; !ok {
		modelInfos[name] = NewModelInfo(db, model)
	}
	return modelInfos[name]
}
