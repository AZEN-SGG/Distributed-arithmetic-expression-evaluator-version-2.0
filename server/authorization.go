package server

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/client"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	defer Close(r)

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var (
		expr = ClientExpression{}
		err  = json.NewDecoder(r.Body).Decode(&expr)
	)

	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if expr.Password == "" || expr.Username == "" {
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	if _, err = client.NewClient(DB, expr.Username, expr.Password); err != nil {
		http.Error(w, fmt.Sprintf("Error registering user: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = fmt.Fprintf(w, "The user under the nickname %s was successfully registered", expr.Username)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	return
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	defer Close(r)
	w.Header().Set("Content-Type", "application/json")
	// Проверяем, что используется метод GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var (
		expr ClientExpression
		err  = json.NewDecoder(r.Body).Decode(&expr)
	)
	if err != nil {
		http.Error(w, decodeErr, http.StatusInternalServerError)
		return
	}

	// Проверяем, что имя пользователя и пароль не пусты
	if expr.Username == "" || expr.Password == "" {
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	// Получаем объект клиента из базы данных
	webUser, err := client.GetClient(DB, expr.Username)
	if err != nil {
		http.Error(w, "Unauthorized: No such user", http.StatusUnauthorized)
		return
	}

	// Генерируем токен для пользователя
	token, err := webUser.GenerateToken()
	if err != nil {
		http.Error(w, "Internal server error while generating token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем статус ответа и отправляем токен
	w.WriteHeader(http.StatusOK)
	if _, err = fmt.Fprint(w, token); err != nil {
		// Если возникает ошибка при отправке токена, логируем её
		// В этот момент изменить статус ответа уже нельзя, поэтому только логируем ошибку
		log.Printf("Failed to write token to response: %v", err)
	}
}

func AuthorizationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Cannot read the request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset r.Body to be readable again

		var expr ClientExpression
		if err := json.Unmarshal(bodyBytes, &expr); err != nil {
			http.Error(w, "Invalid JSON data: "+err.Error(), http.StatusBadRequest)
			return
		}

		if expr.Username == "" || expr.Token == "" {
			http.Error(w, "Username and token cannot be empty", http.StatusBadRequest)
			return
		}

		webUser, err := client.GetClient(DB, expr.Username)
		if err != nil {
			http.Error(w, "Unauthorized - User not found: "+err.Error(), http.StatusUnauthorized)
			return
		}

		err = webUser.VerifyToken(expr.Token)
		if err != nil {
			http.Error(w, "Unauthorized - Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Re-assign the readable body back to the request
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)
	})
}

func ResultHandler(w http.ResponseWriter, r *http.Request) {
	defer Close(r)
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
		return
	}

	var (
		expr = ClientExpression{}
		err  = json.NewDecoder(r.Body).Decode(&expr)
	)
	if err != nil {
		http.Error(w, decodeErr, http.StatusInternalServerError)
		return
	}

	if expr.Username == "" || expr.ID == "" {
		w.WriteHeader(400)
		return
	}

	WebClients.Mu.Lock()
	var webClient = WebClients.Names[expr.Username]
	WebClients.Mu.Unlock()
	if webClient == nil {
		w.WriteHeader(400)
		return
	}

	result, err := webClient.Expressions.GetExpression(expr.ID)
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
