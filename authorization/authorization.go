package authorization

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/client"
	dtbs "Distributed-arithmetic-expression-evaluator-version-2.0/database"
	"fmt"
	"log"
	"net/http"
)

var (
	DB *dtbs.DB
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
		return
	}

	var (
		name     = r.FormValue("username")
		password = r.FormValue("password")
	)

	if name == "" || password == "" {
		w.WriteHeader(400)
		return
	}

	_, err := client.NewClient(DB, name, password)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	_, err = fmt.Fprintf(w, "The user under the nickname %s was successfully registered", name)
	w.WriteHeader(200)
	return
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что используется метод GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем имя пользователя и пароль из запроса
	name := r.FormValue("username")
	password := r.FormValue("password")

	// Проверяем, что имя пользователя и пароль не пусты
	if name == "" || password == "" {
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	// Получаем объект клиента из базы данных
	webUser, err := client.GetClient(DB, name)
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
		var (
			name        = r.FormValue("username")
			password    = r.FormValue("password")
			tokenString = r.FormValue("token")
		)

		if name == "" || password == "" || tokenString == "" {
			w.WriteHeader(400)
			return
		}

		webUser, err := client.GetClient(DB, name)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		err = webUser.VerifyToken(tokenString)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		next.ServeHTTP(w, r)
	})
}
