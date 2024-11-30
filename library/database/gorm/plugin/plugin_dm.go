package plugin

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
)

const (
	DM_OPER_BEFORE_CREATE_NAME = "dm:before_create"
	DM_OPER_BEFORE_UPDATE_NAME = "dm:before_update"
	DM_OPER_BEFORE_QUERY_NAME  = "dm:before_query"
	DM_OPER_BEFORE_DELETE_NAME = "dm:before_delete"
)

type DmPlugin struct {
}

func NewDmPlugin() gorm.Plugin {
	return &DmPlugin{}
}

func (d *DmPlugin) Name() string {
	return "DmPlugin"
}

func (d *DmPlugin) Initialize(db *gorm.DB) (err error) {
	for _, e := range []error{
		//db.Callback().Query().Before("gorm:query").Register(DM_OPER_BEFORE_QUERY_NAME, d.beforeQuery),
		db.Callback().Create().Before("gorm:create").Register(DM_OPER_BEFORE_CREATE_NAME, d.beforeCreate),
		//db.Callback().Update().Before("gorm:update").Register(DM_OPER_BEFORE_UPDATE_NAME, d.beforeUpdate),
		//db.Callback().Delete().Before("gorm:delete").Register(DM_OPER_BEFORE_DELETE_NAME, d.beforeDelete),
	} {
		if e != nil {
			return e
		}
	}
	return
}

func (d *DmPlugin) beforeCreate(db *gorm.DB) {
	callbacks.BeforeCreate(db)
	sqlStr := db.Statement.SQL.String()
	db.Statement.SQL.Reset()
	db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	/*d.replaceFromSqlChar(db)
	d.replaceWhereSqlChar(db)
	d.replaceGroupBySqlChar(db)
	d.replaceOrderBySqlChar(db)*/
}

func (d *DmPlugin) beforeUpdate(db *gorm.DB) {
	callbacks.BeforeUpdate(db)
	sqlStr := db.Statement.SQL.String()
	db.Statement.SQL.Reset()
	db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	/*d.replaceFromSqlChar(db)
	d.replaceWhereSqlChar(db)
	d.replaceGroupBySqlChar(db)
	d.replaceOrderBySqlChar(db)*/
}

func (d *DmPlugin) beforeDelete(db *gorm.DB) {
	callbacks.BeforeDelete(db)
	sqlStr := db.Statement.SQL.String()
	db.Statement.SQL.Reset()
	db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	/*d.replaceFromSqlChar(db)
	d.replaceWhereSqlChar(db)
	d.replaceGroupBySqlChar(db)
	d.replaceOrderBySqlChar(db)*/
}

// query的方式，其实解析"FROM"Clauses是为空的，需要使用shared的函数ChangeCharByDm替换`为"
func (d *DmPlugin) beforeQuery(db *gorm.DB) {
	callbacks.BuildQuerySQL(db)
	sqlStr := db.Statement.SQL.String()
	db.Statement.SQL.Reset()
	db.Statement.SQL.WriteString(addKeywordsQuotes(sqlStr))
	/*d.replaceFromSqlChar(db)
	d.replaceWhereSqlChar(db)
	d.replaceGroupBySqlChar(db)
	d.replaceOrderBySqlChar(db)*/
}

func addKeywordsQuotes(s string) string {
	s = strings.ReplaceAll(s, "`", "\"")
	s = strings.ReplaceAll(s, ".desc,", ".\"desc\",")
	s = strings.ReplaceAll(s, "`desc`", ".\"desc\"")
	s = strings.ReplaceAll(s, ".domain,", ".\"domain\",")
	s = strings.ReplaceAll(s, ", account", ", \"account\"")
	s = strings.ReplaceAll(s, "as version,", "as \"version\",")
	s = strings.ReplaceAll(s, " scope,", " \"scope\",")
	s = strings.ReplaceAll(s, ") value", ") \"value\"")
	s = strings.ReplaceAll(s, "as count", "as \"count\"")
	s = strings.ReplaceAll(s, "GROUP_CONCAT", "WM_CONCAT")
	s = strings.ReplaceAll(s, " role ", " \"role\" ")
	s = strings.ReplaceAll(s, ") role", ") \"role\"")
	s = strings.ReplaceAll(s, "domain like", "\"domain\" like")

	return s
}

// 达梦数据库不能识别字符`，需要替换为"
// 目前根据业务场景考虑替换子句的有from、where、group by、order by
// 后续根据业务需要再增加其他子句的解析和替换
func (d *DmPlugin) replaceFromSqlChar(db *gorm.DB) {
	if db == nil || db.Statement == nil {
		return
	}
	cs, ok := db.Statement.Clauses["FROM"]
	if !ok {
		return
	}

	if cs.Expression == nil {
		return
	}

	from, ok := cs.Expression.(clause.From)
	if !ok {
		return
	}

	if len(from.Joins) == 0 {
		return
	}

	for joinK, join := range from.Joins {
		clauseNamedExpr, ok := join.Expression.(clause.NamedExpr)
		if ok {
			clauseNamedExpr.SQL = strings.Replace(clauseNamedExpr.SQL, "`", "\"", -1)
			from.Joins[joinK].Expression = clauseNamedExpr
		}

		for onK, expr := range join.ON.Exprs {
			clauseExpr, ok := expr.(clause.Expr)
			if !ok {
				continue
			}
			clauseExpr.SQL = strings.Replace(clauseExpr.SQL, "`", "\"", -1)
			from.Joins[joinK].ON.Exprs[onK] = clauseExpr
		}
	}

	cs.Expression = from
	db.Statement.Clauses["FROM"] = cs
}

func (d *DmPlugin) replaceWhereSqlChar(db *gorm.DB) {
	if db == nil || db.Statement == nil {
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

	for k, expr := range where.Exprs {
		clauseExpr, ok := expr.(clause.Expr)
		if !ok {
			continue
		}
		clauseExpr.SQL = strings.Replace(clauseExpr.SQL, "`", "\"", -1)
		where.Exprs[k] = clauseExpr
	}

	cs.Expression = where
	db.Statement.Clauses["WHERE"] = cs
}

func (d *DmPlugin) replaceGroupBySqlChar(db *gorm.DB) {
	if db == nil || db.Statement == nil {
		return
	}
	cs, ok := db.Statement.Clauses["GROUP BY"]
	if !ok {
		return
	}

	if cs.Expression == nil {
		return
	}

	groupBy, ok := cs.Expression.(clause.GroupBy)
	if !ok {
		return
	}

	for k, c := range groupBy.Columns {
		c.Name = strings.Replace(c.Name, "`", "\"", -1)
		groupBy.Columns[k] = c
	}

	for k, expr := range groupBy.Having {
		clauseExpr, ok := expr.(clause.Expr)
		if !ok {
			continue
		}
		clauseExpr.SQL = strings.Replace(clauseExpr.SQL, "`", "\"", -1)
		groupBy.Having[k] = clauseExpr
	}

	cs.Expression = groupBy
	db.Statement.Clauses["GROUP BY"] = cs
}

func (d *DmPlugin) replaceOrderBySqlChar(db *gorm.DB) {
	if db == nil || db.Statement == nil {
		return
	}
	cs, ok := db.Statement.Clauses["ORDER BY"]
	if !ok {
		return
	}

	if cs.Expression == nil {
		return
	}

	orderBy, ok := cs.Expression.(clause.OrderBy)
	if !ok {
		return
	}

	for k, c := range orderBy.Columns {
		c.Column.Name = strings.Replace(c.Column.Name, "`", "\"", -1)
		orderBy.Columns[k] = c
	}

	clauseExpr, ok := orderBy.Expression.(clause.Expr)
	if !ok {
		return
	}
	clauseExpr.SQL = strings.Replace(clauseExpr.SQL, "`", "\"", -1)
	orderBy.Expression = clauseExpr

	cs.Expression = orderBy
	db.Statement.Clauses["ORDER BY"] = cs
}
