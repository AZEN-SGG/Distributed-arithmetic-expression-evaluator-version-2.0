package server

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/client"
	"Distributed-arithmetic-expression-evaluator-version-2.0/database"
	"Distributed-arithmetic-expression-evaluator-version-2.0/rest"
	"log"
	"net/http"
	"strconv"
)

const (
	decodeErr = "Decode JSON data error"
)

var (
	DB         *database.DB
	WebClients *client.Clients
)

type ClientExpression struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	ID       string `json:"id"`
	Content  string `json:"content"`
}

func FormatExpression(id string, expr *rest.Expression) []string {
	var ok, err = expr.GetValue()
	var status string

	switch {
	case err != nil:
		status = err.Error()
	case ok == -1:
		status = "Считается"
	default:
		status = "Высчитан"
	}

	return []string{id, status, expr.Express, expr.Created.Format("02 Jan at 15:04:05"), strconv.FormatInt(expr.Expiration.Milliseconds(), 10) + "ms"}
}

func Close(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Fatal(err)
	}
}
