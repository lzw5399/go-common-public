package dbmodel

import (
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"
)

type DBOptionFunc func(*DBOption)

type DBOption struct {
	TenantTableSet  map[string]struct{}
	I18nTranslators []FieldI18nTranslator
}

func MergeDBOption(opts ...DBOptionFunc) *DBOption {
	option := &DBOption{
		TenantTableSet:  make(map[string]struct{}),
		I18nTranslators: make([]FieldI18nTranslator, 0),
	}
	for _, opt := range opts {
		opt(option)
	}

	return option
}

// WithTenantTables 注册租户表
func WithTenantTables(tables ...interface{}) DBOptionFunc {
	return func(option *DBOption) {
		tableNameSet := make(map[string]struct{}, 1)

		for _, table := range tables {
			v := reflect.ValueOf(table)
			if v.Kind() != reflect.Ptr {
				panic(fmt.Errorf("WithTenantTables value must be a pointer, type(%t)", table))
			}

			// 获取表名
			tabler, ok := table.(schema.Tabler)
			if !ok {
				panic(fmt.Errorf("WithTenantTables value is not Tabler"))
			}
			tableNameSet[tabler.TableName()] = struct{}{}
		}

		fmt.Println("WithTenantTables tableNameSet:", tableNameSet)

		option.TenantTableSet = tableNameSet
	}
}

// WithI18nTables 注册租户表
func WithI18nTables(translators ...FieldI18nTranslator) DBOptionFunc {
	return func(option *DBOption) {
		option.I18nTranslators = translators
	}
}
