package duckdb

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var ErrDuckDBNotSupported = errors.New("DuckDB are not supported this operation")

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

func (m Migrator) CreateTable(values ...interface{}) (err error) {
	if err = m.Migrator.CreateTable(values...); err != nil {
		return
	}
	for _, value := range m.ReorderModels(values, false) {
		if err = m.RunWithValue(value, func(stmt *gorm.Statement) error {
			if stmt.Schema != nil {
				for _, fieldName := range stmt.Schema.DBNames {
					field := stmt.Schema.FieldsByDBName[fieldName]
					if field.Comment != "" {
						if err := m.DB.Exec(
							"COMMENT ON COLUMN ?.? IS ?",
							m.CurrentTable(stmt), clause.Column{Name: field.DBName}, gorm.Expr(m.Migrator.Dialector.Explain("$1", field.Comment)),
						).Error; err != nil {
							return err
						}
					}
				}
			}
			return nil
		}); err != nil {
			return
		}
	}
	return
}

func (m Migrator) DropTable(values ...interface{}) error {
	values = m.ReorderModels(values, false)
	tx := m.DB.Session(&gorm.Session{})
	for i := len(values) - 1; i >= 0; i-- {
		if err := m.RunWithValue(values[i], func(stmt *gorm.Statement) error {
			return tx.Exec("DROP TABLE IF EXISTS ? CASCADE", m.CurrentTable(stmt)).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m Migrator) CurrentSchema(stmt *gorm.Statement, table string) (interface{}, interface{}) {
	if strings.Contains(table, ".") {
		if tables := strings.Split(table, `.`); len(tables) == 2 {
			return tables[0], tables[1]
		}
	}

	if stmt.TableExpr != nil {
		if tables := strings.Split(stmt.TableExpr.SQL, `"."`); len(tables) == 2 {
			return strings.TrimPrefix(tables[0], `"`), table
		}
	}
	return clause.Expr{SQL: "CURRENT_SCHEMA()"}, table
}

func (m Migrator) HasTable(value interface{}) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		currentSchema, curTable := m.CurrentSchema(stmt, stmt.Table)
		return m.DB.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ? AND table_type = ?", currentSchema, curTable, "BASE TABLE").Row().Scan(&count)
	})
	return count > 0
}

func (m Migrator) RenameTable(oldName, newName interface{}) (err error) {
	resolveTable := func(name interface{}) (result string, err error) {
		if v, ok := name.(string); ok {
			result = v
		} else {
			stmt := &gorm.Statement{DB: m.DB}
			if err = stmt.Parse(name); err == nil {
				result = stmt.Table
			}
		}
		return
	}

	var oldTable, newTable string

	if oldTable, err = resolveTable(oldName); err != nil {
		return
	}

	if newTable, err = resolveTable(newName); err != nil {
		return
	}

	if !m.HasTable(oldTable) {
		return
	}

	return m.DB.Exec("RENAME TABLE ? TO ?",
		clause.Table{Name: oldTable},
		clause.Table{Name: newTable},
	).Error
}

func (m Migrator) GetTables() (tableList []string, err error) {
	currentSchema, _ := m.CurrentSchema(m.DB.Statement, "")
	return tableList, m.DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = ?", currentSchema, "BASE TABLE").Row().Scan(&tableList)
}

// Columns

func (m Migrator) AddColumn(value interface{}, field string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if f := stmt.Schema.LookUpField(field); f != nil {
			// avoid using the same name field
			if stmt.Schema == nil {
				return errors.New("failed to get schema")
			}
			f := stmt.Schema.LookUpField(field)
			if f == nil {
				return fmt.Errorf("failed to look up field with name: %s", field)
			}
			if !f.IgnoreMigration {
				return m.DB.Exec(
					"ALTER TABLE ? %s ADD COLUMN ? ?",
					m.CurrentTable(stmt), clause.Column{Name: f.DBName}, m.DB.Migrator().FullDataTypeOf(f),
				).Error
			}
			if f.Comment != "" {
				if err := m.DB.Exec(
					"COMMENT ON COLUMN ?.? IS ?",
					m.CurrentTable(stmt), clause.Column{Name: f.DBName}, gorm.Expr(m.Migrator.Dialector.Explain("$1", f.Comment)),
				).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (m Migrator) DropColumn(dst interface{}, field string) error {
	if err := m.Migrator.DropColumn(dst, field); err != nil {
		return err
	}

	m.resetPreparedStmts()
	return nil
}

// should reset prepared stmts when table changed
// https://duckdb.org/docs/sql/query_syntax/prepared_statements.html
func (m Migrator) resetPreparedStmts() {
	if m.DB.PrepareStmt {
		if pdb, ok := m.DB.ConnPool.(*gorm.PreparedStmtDB); ok {
			pdb.Reset()
		}
	}
}

func (m Migrator) MigrateColumn(value interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	// skip primary field
	if !field.PrimaryKey {
		if err := m.Migrator.MigrateColumn(value, field, columnType); err != nil {
			return err
		}
	}

	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		var description string
		currentSchema, curTable := m.CurrentSchema(stmt, stmt.Table)
		values := []interface{}{currentSchema, curTable, field.DBName, stmt.Table, currentSchema}
		checkSQL := "SELECT description FROM pg_catalog.pg_description "
		checkSQL += "WHERE objsubid = (SELECT ordinal_position FROM information_schema.columns WHERE table_schema = ? AND table_name = ? AND column_name = ?) "
		checkSQL += "AND objoid = (SELECT oid FROM pg_catalog.pg_class WHERE relname = ? AND relnamespace = "
		checkSQL += "(SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = ?))"
		m.DB.Raw(checkSQL, values...).Row().Scan(&description)

		comment := strings.Trim(field.Comment, "'")
		comment = strings.Trim(comment, `"`)
		if field.Comment != "" && comment != description {
			if err := m.DB.Exec(
				"COMMENT ON COLUMN ?.? IS ?",
				m.CurrentTable(stmt), clause.Column{Name: field.DBName}, gorm.Expr(m.Migrator.Dialector.Explain("$1", field.Comment)),
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// TODO: Implement below function.
// AlterColumn(dst interface{}, field string) error

func (m Migrator) HasColumn(value interface{}, field string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		name := field
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				name = field.DBName
			}
		}

		currentSchema, curTable := m.CurrentSchema(stmt, stmt.Table)
		return m.DB.Raw(
			"SELECT count(*) FROM INFORMATION_SCHEMA.columns WHERE table_schema = ? AND table_name = ? AND column_name = ?",
			currentSchema, curTable, name,
		).Scan(&count).Error
	})

	return count > 0
}

func (m Migrator) RenameColumn(dst interface{}, oldName, field string) error {
	if err := m.Migrator.RenameColumn(dst, oldName, field); err != nil {
		return err
	}

	m.resetPreparedStmts()
	return nil
}

// TODO: Implement below function.
// ColumnTypes(dst interface{}) ([]ColumnType, error)

// Views
func (m Migrator) CreateView(name string, option gorm.ViewOption) error {
	return ErrDuckDBNotSupported
}

func (m Migrator) DropView(name string) error {
	return ErrDuckDBNotSupported
}

// // Constraints
// CreateConstraint(dst interface{}, name string) error
// DropConstraint(dst interface{}, name string) error
// HasConstraint(dst interface{}, name string) bool

// // Indexes
// CreateIndex(dst interface{}, name string) error
// DropIndex(dst interface{}, name string) error
// HasIndex(dst interface{}, name string) bool
// RenameIndex(dst interface{}, oldName, newName string) error
