package database

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/rest"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

// TODO: Add test with one time start of NewDB and NewExpressionsDB

const name = "test.db"

func cleanUp(db *DB, t *testing.T) {
	err := db.Close()

	if err != nil {
		t.Error(err)
	}

	err = os.Remove(name)

	if err != nil {
		t.Error(err)
	}
}

func checkColumns(tx *sql.Tx, tableName string, columns []string) error {
	// Получение списка столбцов в таблице
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return err
	}

	existingColumns := make(map[string]bool)
	var (
		cid        int
		columnName string
		ctype      string
		notnull    int
		dfltValue  sql.NullString
		pk         int
	)
	for rows.Next() {
		if err = rows.Scan(&cid, &columnName, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		existingColumns[columnName] = true
	}

	// Проверка наличия каждого требуемого столбца
	for _, col := range columns {
		if _, found := existingColumns[col]; !found {
			return fmt.Errorf("column %s does not exist in table %s", col, tableName)
		}
	}

	return nil
}

func TestNewDB(t *testing.T) {
	db, err := NewDB(name)
	if err != nil {
		t.Error(err)
	}

	defer cleanUp(db, t)

	tx, err := db.Connection.Begin()

	if err != nil {
		t.Error(err)
	}

	err = checkColumns(tx, "users", []string{"name", "password", "secret"})

	if err != nil {
		t.Error(err)
	}

	err = tx.Commit()

	if err != nil {
		t.Error(err)
	}
}

func TestNewExpressionsDB(t *testing.T) {
	db, err := NewExpressionsDB(name)
	if err != nil {
		t.Error(err)
	}

	defer cleanUp(db, t)

	tx, err := db.Connection.Begin()

	if err != nil {
		t.Error(err)
	}

	err = checkColumns(tx, "expressions", []string{"id", "expression", "value", "user", "date"})

	if err != nil {
		t.Error(err)
	}

	err = tx.Commit()

	if err != nil {
		t.Error(err)
	}
}

func TestDB_GetUser(t *testing.T) {
	db, err := NewDB(name)

	if err != nil {
		t.Fatal(err)
	}

	defer cleanUp(db, t)

	_, err = db.CreateUser("name", "password")

	if err != nil {
		t.Fatal(err)
	}

	_, err = db.GetUser("name")

	if err != nil {
		t.Fatal(err)
	}

	_, err = db.GetUser("wrongName")

	if err == nil {
		t.Fatal("There is no user with the name 'wrongName'")
	}
}

func TestDB_ContainsUser(t *testing.T) {
	db, err := NewDB(name)

	if err != nil {
		t.Fatal(err)
	}

	defer cleanUp(db, t)

	_, err = db.CreateUser("name", "password")

	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ContainsUser("name")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ContainsUser("wrongName")
	if err == nil {
		t.Fatal("There is no user with the name 'wrongName'")
	}
}

func TestDB_GetExpression(t *testing.T) {
	db, err := NewExpressionsDB(name)

	if err != nil {
		t.Fatal(err)
	}

	defer cleanUp(db, t)

	expr := &rest.Expression{
		Express: "1+1",
		Value:   2,
		Created: time.Now(),
	}

	err = db.AddExpression(expr, "1", "name")

	if err != nil {
		t.Fatal(err)
	}

	newExpr, err := db.GetExpression("1", "name")

	if err != nil {
		t.Fatal(err)
	}

	if expr.Express != newExpr.Express || expr.Value != newExpr.Value || expr.Created.Sub(newExpr.Created) > time.Millisecond {
		t.Fatal("Expressions are not equal")
	}
}

func Contains(expr []*rest.Expression, el *rest.Expression) bool {
	for _, e := range expr {
		if e.Express == el.Express && e.Value == el.Value && e.Created.Sub(el.Created) < time.Millisecond {
			return true
		}
	}

	return false
}

func TestDB_GetExpressions(t *testing.T) {
	db, err := NewExpressionsDB(name)

	if err != nil {
		t.Fatal(err)
	}

	defer cleanUp(db, t)

	expr := []*rest.Expression{&rest.Expression{
		Express: "1+1",
		Value:   2,
		Created: time.UnixMilli(time.Now().UnixMilli()),
	},
		&rest.Expression{
			Express: "2+2",
			Value:   4,
			Created: time.UnixMilli(time.Now().UnixMilli()),
		},
	}

	for i, el := range expr {
		err = db.AddExpression(el, strconv.Itoa(i), "name")
		if err != nil {
			t.Fatal(err)
		}
	}

	newExpr, err := db.GetExpressions("name")
	if err != nil {
		t.Fatal(err)
	} else if len(newExpr.IDs) == 0 {
		t.Fatal("There is no expressions")
	}

	for _, el := range newExpr.GetExpressions() {
		if ok := Contains(expr, el); !ok {
			t.Fatal("Expressions are not equal")
		}
	}
}
