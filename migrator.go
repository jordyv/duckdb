package duckdb

import (
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Migrator struct {
	migrator.Migrator
	Dialector
}

// Database

func (m Migrator) CurrentDatabase() (name string) {
	m.DB.Raw("SELECT CURRENT_DATABASE()").Row().Scan(&name)
	return
}

func (m Migrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	expr := m.Migrator.FullDataTypeOf(field)

	if value, ok := field.TagSettings["COMMENT"]; ok {
		expr.SQL += " COMMENT " + m.Dialector.Explain("?", value)
	}

	return expr
}

// Tables
CreateTable(dst ...interface{}) error
DropTable(dst ...interface{}) error
HasTable(dst interface{}) bool
RenameTable(oldName, newName interface{}) error
GetTables() (tableList []string, err error)

// Columns
AddColumn(dst interface{}, field string) error
DropColumn(dst interface{}, field string) error
AlterColumn(dst interface{}, field string) error
MigrateColumn(dst interface{}, field *schema.Field, columnType ColumnType) error
HasColumn(dst interface{}, field string) bool
RenameColumn(dst interface{}, oldName, field string) error
ColumnTypes(dst interface{}) ([]ColumnType, error)

// Views
CreateView(name string, option ViewOption) error
DropView(name string) error

// Constraints
CreateConstraint(dst interface{}, name string) error
DropConstraint(dst interface{}, name string) error
HasConstraint(dst interface{}, name string) bool

// Indexes
CreateIndex(dst interface{}, name string) error
DropIndex(dst interface{}, name string) error
HasIndex(dst interface{}, name string) bool
RenameIndex(dst interface{}, oldName, newName string) error
