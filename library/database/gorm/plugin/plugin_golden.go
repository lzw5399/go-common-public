package plugin

import (
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GoldenDbPlugin 解决NOT NULL字段不允许传空字符串
// 新增的时候，默认传" "， 查询的时候，把" "再替换回""
// 更新的时候，如果更新为""， 则改为" "
type GoldenDbPlugin struct {
}

const (
	goldenBeforeCreate = "golden:before_create"
	goldenAfterQuery   = "golden:after_query"
	goldenBeforeUpdate = "golden:before_update"
	goldenBeforeQuery  = "golden:before_query"
)

func NewGoldenDbPlugin() gorm.Plugin {
	return &GoldenDbPlugin{}
}

func (r *GoldenDbPlugin) Name() string {
	return "GoldenDbPlugin"
}

func (r *GoldenDbPlugin) Initialize(db *gorm.DB) (err error) {
	for _, e := range []error{
		db.Callback().Create().Before("gorm:create").Register(goldenBeforeCreate, r.beforeCreate),
		db.Callback().Update().Before("gorm:update").Register(goldenBeforeUpdate, r.beforeUpdate),
		db.Callback().Query().After("gorm:query").Register(goldenAfterQuery, r.afterQuery),
		db.Callback().Query().Before("gorm:query").Register(goldenBeforeQuery, r.beforeQuery),
	} {
		if e != nil {
			return e
		}
	}
	return
}

func (r *GoldenDbPlugin) afterQuery(db *gorm.DB) {
	r.updateDest(db, " ", "")
}

func (r *GoldenDbPlugin) beforeCreate(db *gorm.DB) {
	r.updateDest(db, "", " ")
}

func (r *GoldenDbPlugin) beforeUpdate(db *gorm.DB) {
	r.updateDest(db, "", " ")
}

func (r *GoldenDbPlugin) beforeQuery(db *gorm.DB) {
	r.updateWhereVars(db, "", " ")
}

func (r *GoldenDbPlugin) updateDest(db *gorm.DB, from, to string) {
	if db == nil || db.Statement == nil || db.Statement.Context == nil || db.Statement.Table == "" {
		return
	}

	dest := db.Statement.Dest
	if dest == nil {
		return
	}

	dType := reflect.TypeOf(dest)
	dVal := reflect.ValueOf(dest)
	for dType.Kind() == reflect.Ptr {
		dType = dType.Elem()
		dVal = dVal.Elem()
	}

	if dType.Kind() == reflect.Struct {
		updateStructDefaultStringVal(dest, from, to)
		return
	}

	if dType.Kind() == reflect.Slice || dType.Kind() == reflect.Array {
		l := dVal.Len()
		for i := 0; i < l; i++ {
			value := dVal.Index(i) // Value of item
			typel := value.Type()  // Type of ite
			if typel.Kind() == reflect.Ptr {
				typel = typel.Elem()
			}
			if typel.Kind() != reflect.Struct {
				continue
			}
			updateStructDefaultStringVal(value.Interface(), from, to)
		}
	}

	if dType.Kind() == reflect.Map {
		keys := dVal.MapKeys()
		for _, k := range keys {
			value := dVal.MapIndex(k)
			if value.IsValid() && reflect.TypeOf(value.Interface()).Kind() == reflect.String && reflect.ValueOf(value.Interface()).String() == from {
				dVal.SetMapIndex(k, reflect.ValueOf(to))
			}
		}
	}
}

// 替换struct里面所有字符串
func updateStructDefaultStringVal(dest interface{}, from, to string) {
	sVal := reflect.ValueOf(dest)
	sType := reflect.TypeOf(dest)
	for sType.Kind() == reflect.Ptr {
		sVal = sVal.Elem()
		sType = sType.Elem()
	}
	num := sVal.NumField()
	for i := 0; i < num; i++ {
		valK := sVal.Field(i)
		if valK.Kind() == reflect.String && sVal.Field(i).String() == from {
			sVal.Field(i).SetString(to)
			continue
		}

		if valK.Kind() == reflect.Struct {
			numInner := valK.NumField()
			for j := 0; j < numInner; j++ {
				valKInner := sVal.Field(i).Field(j)
				if valKInner.Kind() == reflect.String && valKInner.String() == from {
					sVal.Field(i).Field(j).SetString(to)
					continue
				}
			}
		}
	}
}

func (r *GoldenDbPlugin) updateWhereVars(db *gorm.DB, from, to string) {
	if db == nil || db.Statement == nil || db.Statement.Context == nil || db.Statement.Table == "" {
		return
	}

	cs, ok := db.Statement.Clauses["WHERE"]
	if !ok {
		return
	}
	if cs.Expression == nil {
		return
	}
	where, ok := cs.Expression.(clause.Where)
	if !ok {
		return
	}
	for _, expr := range where.Exprs {
		if expr == nil {
			continue
		}
		clauseExpr, ok := expr.(clause.Expr)
		if !ok {
			continue
		}
		for i, v := range clauseExpr.Vars {
			if v == from {
				clauseExpr.Vars[i] = to
			}
		}
	}
}
