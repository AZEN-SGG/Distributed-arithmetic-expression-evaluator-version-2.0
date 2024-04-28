package database

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/expressions"
	"Distributed-arithmetic-expression-evaluator-version-2.0/rest"
	"crypto/rand"
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
	"log"
	"math/big"
	"time"
)

type DB struct {
	Connection *sql.DB
}

type DBUser struct {
	Name     string
	Password string
	Secret   string
}

// CreateDataBase creates a database either by the first arg or by default
func CreateDataBase(db *sql.DB, args ...interface{}) error {
	var createStmt string

	// Begin a transaction.
	tx, err := db.Begin()

	if err != nil {
		return err
	}

	if args != nil {
		createStmt = args[0].(string)
	} else {
		createStmt = `
    	CREATE TABLE IF NOT EXISTS users (
        name TEXT PRIMARY KEY,
        password TEXT NOT NULL,
        secret TEXT NOT NULL
    );`
	}

	// Execute the SQL statement.
	_, err = tx.Exec(createStmt)
	if err != nil {
		err = tx.Rollback() // Roll back the transaction on error.

		if err != nil {
			return err
		}

		return err
	}

	// Commit the transaction to finalize changes.
	return tx.Commit()
}

// NewDB creates a new DB object with connection to the database
func NewDB(name string, args ...interface{}) (*DB, error) {
	db, err := sql.Open("sqlite3", name)

	if err != nil {
		return nil, err
	}

	err = CreateDataBase(db, args...)

	if err != nil {
		return nil, err
	}

	return &DB{
		Connection: db,
	}, nil
}

// NewExpressionsDB is a copy of NewDB, but with the table installed
func NewExpressionsDB(name string) (*DB, error) {
	var arg = `CREATE TABLE IF NOT EXISTS expressions (id TEXT, expression TEXT, value INT, user TEXT, date INT, PRIMARY KEY (id, user));`

	var db, err = NewDB(name, arg)

	if err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.Connection.Close()
}

func Close(tx *sql.Tx) {
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

// RandomNumber generate a random number
func RandomNumber(num, from, to int64) (*big.Int, error) {
	low := new(big.Int).Exp(big.NewInt(num), big.NewInt(from), nil)
	high := new(big.Int).Exp(big.NewInt(num), big.NewInt(to), nil)
	rangeNum := new(big.Int).Sub(high, low)

	randomNum, err := rand.Int(rand.Reader, rangeNum)

	if err != nil {
		return nil, err
	}

	randomNum = randomNum.Add(randomNum, low)
	return randomNum, nil
}

// CreateUser generate a secret, after that it adds a new user to the database
func (db *DB) CreateUser(name, password string) (*DBUser, error) {
	var randomNum, err = RandomNumber(2, 127, 128)

	if err != nil {
		return nil, err
	}

	AddUserStmt := `INSERT INTO users (name, password, secret) VALUES ($1, $2, $3);`

	tx, err := db.Connection.Begin()

	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(AddUserStmt, name, password, randomNum.String())

	if err != nil {
		anotherErr := tx.Rollback()

		if anotherErr != nil {
			return nil, anotherErr
		}

		return nil, err
	}

	return &DBUser{
		Name:     name,
		Password: password,
		Secret:   randomNum.String(),
	}, tx.Commit()
}

// GetUser retrieve a user from the database
func (db *DB) GetUser(name string) (*DBUser, error) {
	var (
		err     error
		tx      *sql.Tx
		getStmt = `SELECT * FROM users WHERE name = $1;`
	)

	tx, err = db.Connection.Begin()

	if err != nil {
		return nil, err
	}

	defer Close(tx)

	var user DBUser
	err = tx.QueryRow(getStmt, name).Scan(&user.Name, &user.Password, &user.Secret)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetUsers() ([]*DBUser, error) {
	var (
		getStmt = `SELECT * FROM users;`
		tx, err = db.Connection.Begin()
	)
	if err != nil {
		return nil, err
	}

	defer Close(tx)
	rows, err := tx.Query(getStmt)
	if err != nil {
		return nil, err
	}

	var (
		users = []*DBUser{}
		user  DBUser
	)
	for rows.Next() {
		err = rows.Scan(&user.Name, &user.Password, &user.Secret)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (db *DB) ContainsUser(name string) (bool, error) {
	var getStmt = `SELECT name FROM users WHERE name = $1;`
	tx, err := db.Connection.Begin()
	if err != nil {
		return false, err
	}

	defer Close(tx)

	row := tx.QueryRow(getStmt, name)
	var Name string
	row.Scan(&Name)
	if Name == "" {
		return false, errors.New("user not found")
	}

	return true, nil
}

func (db *DB) AddExpression(expr *rest.Expression, id, user string) error {
	var addStmt = `INSERT INTO expressions (id, expression, value, user, date) VALUES ($1, $2, $3, $4, $5);`
	tx, err := db.Connection.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec(addStmt, id, expr.Express, expr.Value, user, expr.Created.UnixMilli())

	if err != nil {
		anErr := tx.Rollback()

		if anErr != nil {
			return anErr
		}

		return err
	}

	return tx.Commit()
}

func (db *DB) ChangeExpression(value int, id, user string) error {
	var changeStmt = `UPDATE expressions SET value = $2 WHERE id = $4 AND user = $5;`
	tx, err := db.Connection.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(changeStmt, value, id, user)
	if err != nil {
		anErr := tx.Rollback()
		if anErr != nil {
			return anErr
		}
		return err
	}

	return tx.Commit()
}

func (db *DB) GetExpression(id, userName string) (*rest.Expression, error) {
	var (
		getStmt  = `SELECT expression, value, date FROM expressions WHERE id = $1 AND user = $2;`
		unixTime int64
		expr     rest.Expression
		err      = db.Connection.QueryRow(getStmt, id, userName).Scan(&expr.Express, &expr.Value, &unixTime)
	)

	if err != nil {
		return nil, err
	}

	expr.Created = time.UnixMilli(unixTime)

	return &expr, nil
}

func (db *DB) GetExpressions(userName string) (*expressions.Expressions, error) {
	var (
		expresses = expressions.NewExpressions()
		GetStmt   = `SELECT id, expression, value, date FROM expressions WHERE user = $1;`
		err       error
		tx        *sql.Tx
		rows      *sql.Rows
	)

	tx, err = db.Connection.Begin()
	if err != nil {
		return nil, err
	}

	defer Close(tx)

	rows, err = tx.Query(GetStmt, userName)
	if err != nil {
		return nil, err
	}

	var (
		id      string
		expr    string
		value   string
		created int64
	)
	for rows.Next() {
		err = rows.Scan(&id, &expr, &value, &created)
		if err != nil {
			return nil, err
		}
		_, err = expresses.AddExpression(id, expr, time.UnixMilli(created), value)
		if err != nil {
			return nil, err
		}
	}

	return expresses, nil
}
