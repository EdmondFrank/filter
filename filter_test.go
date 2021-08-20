package filter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils/tests"
)

func TestSQLEscape(t *testing.T) {
	tx := &gorm.DB{Config: &gorm.Config{
		Dialector: tests.DummyDialector{},
	}}
	assert.Equal(t, "`name`", SQLEscape(tx, "name"))
}

func TestGetTableName(t *testing.T) {
	type testModel struct {
		Name string
		ID   uint
	}

	tx := &gorm.DB{
		Config:    &gorm.Config{Dialector: tests.DummyDialector{}},
		Statement: &gorm.Statement{},
	}
	tx.Statement.DB = tx

	assert.Empty(t, getTableName(tx))

	tx = tx.Table("users")

	assert.Equal(t, "users.", getTableName(tx))

	tx, _ = gorm.Open(tests.DummyDialector{}, nil)
	tx = tx.Model(&testModel{})

	assert.Equal(t, "test_models.", getTableName(tx))

	assert.Panics(t, func() {
		tx, _ = gorm.Open(tests.DummyDialector{}, nil)
		tx = tx.Model(1)
		fmt.Println(getTableName(tx))
	})
}

func TestFilterWhere(t *testing.T) {
	db, _ := gorm.Open(&tests.DummyDialector{}, nil)
	filter := &Filter{Field: "name", Args: []string{"val1"}}
	db = filter.Where(db, "name = ?", "val1")
	expected := map[string]clause.Clause{
		"WHERE": {
			Name: "WHERE",
			Expression: clause.Where{
				Exprs: []clause.Expression{
					clause.Expr{SQL: "name = ?", Vars: []interface{}{"val1"}},
				},
			},
		},
	}
	assert.Equal(t, expected, db.Statement.Clauses)
}

func TestFilterWhereOr(t *testing.T) {
	db, _ := gorm.Open(&tests.DummyDialector{}, nil)
	filter := &Filter{Field: "name", Args: []string{"val1"}, Or: true}
	db = filter.Where(db, "name = ?", "val1")
	expected := map[string]clause.Clause{
		"WHERE": {
			Name: "WHERE",
			Expression: clause.Where{
				Exprs: []clause.Expression{
					clause.OrConditions{
						Exprs: []clause.Expression{
							clause.Expr{SQL: "name = ?", Vars: []interface{}{"val1"}},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, db.Statement.Clauses)
}
