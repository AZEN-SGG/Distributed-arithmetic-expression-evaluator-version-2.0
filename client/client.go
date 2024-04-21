package client

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/database"
	"Distributed-arithmetic-expression-evaluator-version-2.0/expressions"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"sync"
)

// Client структура представляет клиента системы с его личными данными и выражениями.
type Client struct {
	name        string                   // имя клиента
	password    string                   // пароль клиента
	secret      string                   // секрет для генерации JWT токена
	Expressions *expressions.Expressions // коллекция выражений, связанных с клиентом
}

// NewClient создает новый экземпляр клиента, проверяя, что обязательные поля не пустые.
func NewClient(db *database.DB, name, password string) (*Client, error) {
	if name == "" || password == "" {
		return nil, errors.New("name, password and secret cannot be empty") // валидация входных данных
	}

	dBUser, err := db.CreateUser(name, password) // создание пользователя
	if err != nil {
		return nil, err
	}

	return &Client{
		name:        name,
		password:    password,
		Expressions: expressions.NewExpressions(), // инициализация новой коллекции выражений
		secret:      dBUser.Secret,
	}, nil
}

func GetClient(db *database.DB, name string) (*Client, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	user, err := db.GetUser(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	client := &Client{
		name:        user.Name,
		password:    user.Password,
		secret:      user.Secret,
		Expressions: expressions.NewExpressions(),
	}

	// Дальнейшая логика
	return client, nil
}

// AddExpression добавляет новое выражение в коллекцию клиента и записывает в базу данных.
func (c *Client) AddExpression(db *database.DB, ID, expr string) error {
	objExpr, err := c.Expressions.AddExpression(ID, expr) // добавление выражения в коллекцию
	if err != nil {
		return err // обработка возможной ошибки
	}

	if err = db.AddExpression(objExpr, ID, c.name); err != nil {
		return err // запись выражения в базу данных и обработка возможной ошибки
	}

	return nil
}

// GenerateToken генерирует JWT токен для клиента.
func (c *Client) GenerateToken() (string, error) {
	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name":     c.name, // использование имени и пароля в claims токена
		"password": c.password,
	})

	var tokenString, err = token.SignedString([]byte(c.secret)) // подпись токена секретом клиента
	if err != nil {
		return "", err // обработка возможной ошибки при подписи
	}

	return tokenString, nil
}

// VerifyToken проверяет валидность переданного токена.
func (c *Client) VerifyToken(tokenString string) error {
	var token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.secret), nil // использование секрета клиента для верификации токена
	})
	if err != nil {
		return err // обработка ошибки разбора токена
	}

	if !token.Valid {
		return fmt.Errorf("invalid token") // возвращение ошибки, если токен недействителен
	}

	return nil
}

type Clients struct {
	Names map[string]*Client
	Mu    sync.Mutex
}

func NewClients(db *database.DB) (*Clients, error) {
	userNames, err := db.GetUsers()
	if err != nil {
		return nil, err
	}

	var (
		names   = make(map[string]*Client)
		webUser *Client
	)

	for _, el := range userNames {
		webUser, err = GetClient(db, el.Name)
		if err != nil {
			return nil, err
		}
		names[el.Name] = webUser // добавление пользователя в коллекцию
	}

	return &Clients{
		Names: names,
		Mu:    sync.Mutex{},
	}, nil
}
