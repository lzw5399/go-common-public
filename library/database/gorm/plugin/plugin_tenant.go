package plugin

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	fcontext "github.com/lzw5399/go-common-public/library/context"
)

type TenantPlugin struct {
	tenantTableSet map[string]struct{}
}

const (
	eventBeforeUpdate = "tenant:before_update"
	eventBeforeQuery  = "tenant:before_query"
	eventBeforeDelete = "tenant:before_delete"
	eventBeforeRow    = "tenant:before_row"
	eventBeforeRaw    = "tenant:before_raw"

	opCreate = "create"
	opUpdate = "update"
	opQuery  = "query"
	opDelete = "delete"
	opRow    = "row"
	opRaw    = "raw"
)

func NewTenantPlugin(tenantTableSet map[string]struct{}) gorm.Plugin {
	return &TenantPlugin{
		tenantTableSet: tenantTableSet,
	}
}

func (i *TenantPlugin) Name() string {
	return "TenantPlugin"
}

func (i *TenantPlugin) Initialize(db *gorm.DB) (err error) {
	// Register various callback events in GORM
	for _, e := range []error{
		// db.Callback().Create().Before("gorm:create").Register(_eventBeforeCreate, beforeCreate),
		db.Callback().Update().Before("gorm:update").Register(eventBeforeUpdate, i.beforeUpdate),
		db.Callback().Query().Before("gorm:query").Register(eventBeforeQuery, i.beforeQuery),
		db.Callback().Delete().Before("gorm:delete").Register(eventBeforeDelete, i.beforeDelete),
		db.Callback().Row().Before("gorm:row").Register(eventBeforeRow, i.beforeRow),
		db.Callback().Raw().Before("gorm:raw").Register(eventBeforeRaw, i.beforeRaw),
	} {
		if e != nil {
			return e
		}
	}
	return
}

//func beforeCreate(db *gorm.DB) {
//	injectBefore(db, opCreate)
//}

func (i *TenantPlugin) beforeUpdate(db *gorm.DB) {
	i.injectBefore(db, opUpdate)
	if fconfig.DefaultConfig.DBMode == fconfig.DB_MODE_DM {
		callbacks.BeforeUpdate(db)
		sqlStr := db.Statement.SQL.String()
		db.Statement.SQL.Reset()
		db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	}
}

func (i *TenantPlugin) beforeQuery(db *gorm.DB) {
	i.injectBefore(db, opQuery)
	if fconfig.DefaultConfig.DBMode == fconfig.DB_MODE_DM {
		callbacks.BuildQuerySQL(db)
		sqlStr := db.Statement.SQL.String()
		db.Statement.SQL.Reset()
		db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	}
}

func (i *TenantPlugin) beforeDelete(db *gorm.DB) {
	i.injectBefore(db, opDelete)
	if fconfig.DefaultConfig.DBMode == fconfig.DB_MODE_DM {
		callbacks.BeforeDelete(db)
		sqlStr := db.Statement.SQL.String()
		db.Statement.SQL.Reset()
		db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	}
}

func (i *TenantPlugin) beforeRow(db *gorm.DB) {
	i.injectBefore(db, opRow)
	if fconfig.DefaultConfig.DBMode == fconfig.DB_MODE_DM {
		callbacks.BuildQuerySQL(db)
		sqlStr := db.Statement.SQL.String()
		db.Statement.SQL.Reset()
		db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	}
}

func (i *TenantPlugin) beforeRaw(db *gorm.DB) {
	i.injectBefore(db, opRaw)
	if fconfig.DefaultConfig.DBMode == fconfig.DB_MODE_DM {
		callbacks.BuildQuerySQL(db)
		sqlStr := db.Statement.SQL.String()
		db.Statement.SQL.Reset()
		db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	}
}

func (i *TenantPlugin) injectBefore(db *gorm.DB, op string) {
	if db == nil || db.Statement == nil || db.Statement.Context == nil || db.Statement.Table == "" {
		return
	}

	// 判断当前表是否是支持多租户的表
	_, ok := i.tenantTableSet[db.Statement.Table]
	if !ok {
		return
	}

	// 判断是否手动跳过租户条件
	ignoreTenant := fcontext.IgnoreTenantFromContext(db.Statement.Context)
	if ignoreTenant {
		return
	}

	// 获取当前用户信息
	currentUser := fcontext.UserInfoFromContext(db.Statement.Context)
	if currentUser == nil {
		return
	}

	if currentUser.OrgId == 0 {
		return
	}

	tenantWhereExpr := fmt.Sprintf("%s.org_id = ?", db.Statement.Table)

	// 获取到 where 子句的expression对象
	cs, ok := db.Statement.Clauses["WHERE"]
	if !ok {
		// 如果没有where子句，则直接添加租户where条件
		db.Where(tenantWhereExpr, currentUser.OrgId)
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

		// 判断是否包含org_id条件，包含则直接返回
		sql := strings.ToLower(clauseExpr.SQL)
		if strings.Contains(sql, "org_id") {
			return
		}
	}

	// 将org_id prepend 到sql的 where 子句中, 方便命中索引
	tenantFilterExprs := make([]clause.Expression, 0, len(where.Exprs)+1)
	tenantFilterExprs = append(tenantFilterExprs, clause.AndConditions{
		Exprs: []clause.Expression{
			gorm.Expr(tenantWhereExpr, currentUser.OrgId),
		},
	})

	where.Exprs = append(tenantFilterExprs, where.Exprs...)
	cs.Expression = where
	db.Statement.Clauses["WHERE"] = cs
}
