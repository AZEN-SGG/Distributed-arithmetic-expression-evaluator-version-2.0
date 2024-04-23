package server

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/calculator"
	"Distributed-arithmetic-expression-evaluator-version-2.0/client"
	"Distributed-arithmetic-expression-evaluator-version-2.0/data"
	"Distributed-arithmetic-expression-evaluator-version-2.0/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func ArithmeticsHandler(w http.ResponseWriter, r *http.Request) {
	defer Close(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var expr ClientExpression
	if err := json.NewDecoder(r.Body).Decode(&expr); err != nil {
		http.Error(w, "Invalid JSON data: "+err.Error(), http.StatusBadRequest)
		return
	}

	if expr.Username == "" || expr.ID == "" || expr.Content == "" {
		http.Error(w, "Username, ID, and content must not be empty", http.StatusBadRequest)
		return
	}

	WebClients.Mu.Lock()
	webClient, exists := WebClients.Names[expr.ID]
	WebClients.Mu.Unlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	preparedContent, err := calculator.PreparingExpression(expr.Content)
	if err != nil {
		http.Error(w, "Error preparing expression: "+err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = webClient.Expressions.AddExpression(expr.ID, preparedContent); err != nil {
		http.Error(w, "Error adding expression: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = fmt.Fprint(w, "Expression added successfully")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ListProcessHandler(w http.ResponseWriter, r *http.Request) {
	defer Close(r)
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

	mux.Handle("/get", AuthorizationMiddleware(ResultHandler))
	mux.Handle("/list", AuthorizationMiddleware(ListProcessHandler))
	mux.Handle("/expression", AuthorizationMiddleware(ArithmeticsHandler))
	mux.HandleFunc("/math", MathOperationsHandler)
	mux.HandleFunc("/processes", ProcessesHandler)
	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("/register", RegisterHandler)

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
