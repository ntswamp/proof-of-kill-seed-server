package lib_db

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
)

/**
 * CustomType向けのUpdateCallback.
 * gormのupdateCallbackに追記しただけ.
 */
func UpdateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var sqls []string
		modelInfo := getModelInfo(scope.DB(), scope.Value)
		if updateAttrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
			// Sort the column names so that the generated SQL is the same every time.
			updateMap := updateAttrs.(map[string]interface{})
			var columns []string
			for c := range updateMap {
				columns = append(columns, c)
			}
			sort.Strings(columns)

			for _, column := range columns {
				customType := modelInfo.GetCustomType(column)
				value := updateMap[column]
				if customType != nil {
					customValue, err := customType.Update(value)
					if err != nil {
						scope.Err(err)
						return
					}
					sqls = append(sqls, fmt.Sprintf("%v = %s", scope.Quote(column), customValue))
				} else {
					sqls = append(sqls, fmt.Sprintf("%v = %v", scope.Quote(column), scope.AddToVars(value)))
				}
			}
		} else {
			for _, field := range scope.Fields() {
				if changeableField(scope, field) {
					// CustumType.
					customType := modelInfo.GetCustomType(field.Name)
					if customType != nil {
						customValue, err := customType.Update(field.Field.Interface())
						if err != nil {
							scope.Err(err)
							return
						}
						sqls = append(sqls, fmt.Sprintf("%v = %s", scope.Quote(field.DBName), customValue))
					} else if !field.IsPrimaryKey && field.IsNormal && (field.Name != "CreatedAt" || !field.IsBlank) {
						if !field.IsForeignKey || !field.IsBlank || !field.HasDefaultValue {
							sqls = append(sqls, fmt.Sprintf("%v = %v", scope.Quote(field.DBName), scope.AddToVars(field.Field.Interface())))
						}
					} else if relationship := field.Relationship; relationship != nil && relationship.Kind == "belongs_to" {
						for _, foreignKey := range relationship.ForeignDBNames {
							if foreignField, ok := scope.FieldByName(foreignKey); ok && !changeableField(scope, foreignField) {
								sqls = append(sqls,
									fmt.Sprintf("%v = %v", scope.Quote(foreignField.DBName), scope.AddToVars(foreignField.Field.Interface())))
							}
						}
					}
				}
			}
		}

		var extraOption string
		if str, ok := scope.Get("gorm:update_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		if len(sqls) > 0 {
			sql := fmt.Sprintf(
				"UPDATE %v SET %v%v%v",
				scope.QuotedTableName(),
				strings.Join(sqls, ", "),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)
			scope.Raw(sql).Exec()
		}
	}
}
