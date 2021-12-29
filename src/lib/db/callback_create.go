package lib_db

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

/**
 * CustomType向けのCreateCallback.
 * gormのcreateCallbackに追記しただけ.
 */
func CreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		modelInfo := getModelInfo(scope.DB(), scope.Value)
		var (
			columns, placeholders        []string
			blankColumnsWithDefaultValue []string
		)
		for _, field := range scope.Fields() {
			if changeableField(scope, field) {
				if field.IsNormal && !field.IsIgnored {
					// CustumType.
					customType := modelInfo.GetCustomType(field.Name)
					if customType != nil {
						customValue, err := customType.Update(field.Field.Interface())
						if err != nil {
							scope.Err(err)
							return
						}
						columns = append(columns, scope.Quote(field.DBName))
						placeholders = append(placeholders, customValue)
					} else if field.IsBlank && field.HasDefaultValue {
						blankColumnsWithDefaultValue = append(blankColumnsWithDefaultValue, scope.Quote(field.DBName))
						scope.InstanceSet("gorm:blank_columns_with_default_value", blankColumnsWithDefaultValue)
					} else if !field.IsPrimaryKey || !field.IsBlank {
						columns = append(columns, scope.Quote(field.DBName))
						placeholders = append(placeholders, scope.AddToVars(field.Field.Interface()))
					}
				} else if field.Relationship != nil && field.Relationship.Kind == "belongs_to" {
					for _, foreignKey := range field.Relationship.ForeignDBNames {
						if foreignField, ok := scope.FieldByName(foreignKey); ok && !changeableField(scope, foreignField) {
							columns = append(columns, scope.Quote(foreignField.DBName))
							placeholders = append(placeholders, scope.AddToVars(foreignField.Field.Interface()))
						}
					}
				}
			}
		}
		var (
			returningColumn = "*"
			quotedTableName = scope.QuotedTableName()
			primaryField    = scope.PrimaryField()
			extraOption     string
			insertModifier  string
		)
		if str, ok := scope.Get("gorm:insert_option"); ok {
			extraOption = fmt.Sprint(str)
		}
		if str, ok := scope.Get("gorm:insert_modifier"); ok {
			insertModifier = strings.ToUpper(fmt.Sprint(str))
			if insertModifier == "INTO" {
				insertModifier = ""
			}
		}
		if primaryField != nil {
			returningColumn = scope.Quote(primaryField.DBName)
		}
		lastInsertIDReturningSuffix := scope.Dialect().LastInsertIDReturningSuffix(quotedTableName, returningColumn)

		if len(columns) == 0 {
			scope.Raw(fmt.Sprintf(
				"INSERT %v INTO %v %v%v%v",
				addExtraSpaceIfExist(insertModifier),
				quotedTableName,
				scope.Dialect().DefaultValueStr(),
				addExtraSpaceIfExist(extraOption),
				addExtraSpaceIfExist(lastInsertIDReturningSuffix),
			))
		} else {
			scope.Raw(fmt.Sprintf(
				"INSERT %v INTO %v (%v) VALUES (%v)%v%v",
				addExtraSpaceIfExist(insertModifier),
				scope.QuotedTableName(),
				strings.Join(columns, ","),
				strings.Join(placeholders, ","),
				addExtraSpaceIfExist(extraOption),
				addExtraSpaceIfExist(lastInsertIDReturningSuffix),
			))
		}

		// execute create sql
		if lastInsertIDReturningSuffix == "" || primaryField == nil {
			if result, err := scope.SQLDB().Exec(scope.SQL, scope.SQLVars...); scope.Err(err) == nil {
				// set rows affected count
				scope.DB().RowsAffected, _ = result.RowsAffected()

				// set primary value to primary field
				if primaryField != nil && primaryField.IsBlank {
					if primaryValue, err := result.LastInsertId(); scope.Err(err) == nil {
						scope.Err(primaryField.Set(primaryValue))
					}
				}
			}
		} else {
			if primaryField.Field.CanAddr() {
				if err := scope.SQLDB().QueryRow(scope.SQL, scope.SQLVars...).Scan(primaryField.Field.Addr().Interface()); scope.Err(err) == nil {
					primaryField.IsBlank = false
					scope.DB().RowsAffected = 1
				}
			} else {
				scope.Err(gorm.ErrUnaddressable)
			}
		}
	}
}
