package server

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/authorization"
	"Distributed-arithmetic-expression-evaluator-version-2.0/calculator"
	"Distributed-arithmetic-expression-evaluator-version-2.0/client"
	"Distributed-arithmetic-expression-evaluator-version-2.0/data"
	"Distributed-arithmetic-expression-evaluator-version-2.0/database"
	"Distributed-arithmetic-expression-evaluator-version-2.0/rest"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	DB         *database.DB
	WebClients *client.Clients
)

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

func ResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
		return
	}

	var (
		name = r.FormValue("username")
		id   = r.FormValue("id")
	)
	if name == "" || id == "" {
		w.WriteHeader(400)
		return
	}

	WebClients.Mu.Lock()
	var webClient = WebClients.Names[id]
	WebClients.Mu.Unlock()
	if webClient == nil {
		w.WriteHeader(400)
		return
	}

	var result, err = webClient.Expressions.GetExpression(id)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	if result.Value == -1 {
		result.Value, err = result.GetValue()

		if err != nil {
			w.WriteHeader(400)
			return
		}
	}

	_, err = fmt.Fprintf(w, "Expression - %s = %d\nCreation data: %s\nTime: %s", result.Express, result.Value, result.Created, result.Expiration)

	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func ListProcessHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		name = r.FormValue("username")
	)
	if name == "" {
		w.WriteHeader(400)
		return
	}
	WebClients.Mu.Lock()
	webClient := WebClients.Names[name]
	WebClients.Mu.Unlock()
	if webClient == nil {
		w.WriteHeader(400)
		return
	}

	_, err = fmt.Fprintln(w, "List of process:")

	if err != nil {
		w.WriteHeader(500)
		return
	}

	_, err = fmt.Fprint(w, "Format: ID - state - expression - creation date - approximate calculation time\n")

	if err != nil {
		w.WriteHeader(500)
		return
	}

	_, err = fmt.Fprintln(w, "")

	if err != nil {
		w.WriteHeader(500)
		return
	}

	var values = webClient.Expressions.GetExpressions()

	for id, expr := range values {
		formatExpression := FormatExpression(id, expr)

		_, err = fmt.Fprint(w, strings.Join(formatExpression, " - ")+"\n")

		if err != nil {
			w.WriteHeader(500)
			return
		}
	}
}

func ArithmeticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
		return
	}
	var (
		expr = strings.Replace(r.FormValue("expression"), " ", "+", -1)
		id   = r.FormValue("id")
		name = r.FormValue("username")
		err  error
	)
	if name == "" || id == "" || expr == "" {
		w.WriteHeader(400)
		return
	}

	WebClients.Mu.Lock()
	var webClient = WebClients.Names[id]
	WebClients.Mu.Unlock()
	if webClient == nil {
		w.WriteHeader(400)
		return
	}

	expr, err = calculator.PreparingExpression(expr)

	if err != nil {
		w.WriteHeader(400)
		return
	}

	_, err = webClient.Expressions.AddExpression(id, expr)

	if err != nil {
		w.WriteHeader(400)
		return
	}

}

func MathOperationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var _, err = fmt.Fprintf(w, "Math operations:\n+ : %s\n- : %s\n* : %s\n/ : %s", calculator.ArithmeticExecTime[43], calculator.ArithmeticExecTime[45], calculator.ArithmeticExecTime[42], calculator.ArithmeticExecTime[47])

		if err != nil {
			w.WriteHeader(500)
			return
		}

	} else if r.Method == http.MethodPost {
		// Обработка POST запроса с обновленными значениями операций
		var err = r.ParseForm()

		if err != nil {
			w.WriteHeader(500)
			return
		}

		addition := r.Form.Get("addition")
		subtraction := r.Form.Get("subtraction")
		multiplication := r.Form.Get("multiplication")
		division := r.Form.Get("division")

		var operations = make([]*calculator.Operation, 0, 4)
		var operation *calculator.Operation

		if addition != "" {
			operation, err = calculator.FormatOperation(43, addition)
			if err != nil {
				w.WriteHeader(400)
				return
			}

			operations = append(operations, operation)
		}

		if subtraction != "" {
			operation, err = calculator.FormatOperation(45, subtraction)

			if err != nil {
				w.WriteHeader(400)
				return
			}

			operations = append(operations, operation)
		}

		if multiplication != "" {
			operation, err = calculator.FormatOperation(42, multiplication)

			if err != nil {
				w.WriteHeader(400)
				return
			}

			operations = append(operations, operation)
		}

		if division != "" {
			operation, err = calculator.FormatOperation(47, division)

			if err != nil {
				w.WriteHeader(400)
				return
			}

			operations = append(operations, operation)
		}

		calculator.MathOperation(operations...)
		err = data.UploadArithmetic(calculator.ArithmeticExecTime, "data/arithmetic.csv")

		if err != nil {
			w.WriteHeader(500)
			return
		}

		// Вывод обновленных значений операций
		_, err = fmt.Fprintf(w, "Operations updated:\n+ : %s\n- : %s\n* : %s\n/ : %s", calculator.ArithmeticExecTime[43], calculator.ArithmeticExecTime[45], calculator.ArithmeticExecTime[42], calculator.ArithmeticExecTime[47])

		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
	} else {
		w.WriteHeader(400)
	}
}

func ProcessesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var _, err = fmt.Fprint(w, "Processes:\n\n")

		if err != nil {
			w.WriteHeader(500)
			return
		}

		for i, elem := range calculator.ComputingPower {
			_, err = fmt.Fprintf(w, "%d - %s\n", i, string(elem))

			if err != nil {
				w.WriteHeader(500)
				return
			}
		}

	} else {
		w.WriteHeader(400)
		return
	}
}

func MuxHandler() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/get", authorization.AuthorizationMiddleware(ResultHandler))
	mux.Handle("/list", authorization.AuthorizationMiddleware(ListProcessHandler))
	mux.Handle("/expression", authorization.AuthorizationMiddleware(ArithmeticsHandler))
	mux.HandleFunc("/math", MathOperationsHandler)
	mux.HandleFunc("/processes", ProcessesHandler)
	mux.HandleFunc("/login", authorization.LoginHandler)
	mux.HandleFunc("/register", authorization.RegisterHandler)
	return mux
}

func StartHandler(port string) {
	var err error
	DB, err = database.NewDB("database/data.db")
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	if DB == nil {
		log.Fatal("Database is nil")
	}

	authorization.DB = DB
	WebClients, err = client.NewClients(DB)
	if err != nil {
		log.Fatal("Failed to create clients: ", err)
	}

	var mux = MuxHandler()
	log.Printf("Server start listening on http://localhost:%s/", port)
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
