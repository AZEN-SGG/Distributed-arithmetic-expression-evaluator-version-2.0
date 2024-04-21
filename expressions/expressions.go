package expressions

import (
	// Импорт зависимостей из других пакетов проекта и стандартных библиотек
	"Distributed-arithmetic-expression-evaluator-version-2.0/calculator"
	"Distributed-arithmetic-expression-evaluator-version-2.0/data"
	"Distributed-arithmetic-expression-evaluator-version-2.0/rest"
	"errors"
	"slices"
	"strconv"
	"sync"
	"time"
)

// Expressions структура для управления коллекцией арифметических выражений.
type Expressions struct {
	IDs map[string]*rest.Expression // Мапа, связывающая ID с объектами Expression
	mu  sync.Mutex                  // Мьютекс для синхронизации доступа к мапе
}

// NewExpressions создает и возвращает новый экземпляр структуры Expressions.
func NewExpressions() *Expressions {
	return &Expressions{
		IDs: map[string]*rest.Expression{},
		mu:  sync.Mutex{},
	}
}

// AddExpression добавляет новое выражение в коллекцию.
func (express *Expressions) AddExpression(ID, expr string, args ...interface{}) (*rest.Expression, error) {
	express.mu.Lock() // Блокировка для безопасного доступа к мапе
	var keys = rest.MapGetKeys(express.IDs)
	express.mu.Unlock() // Разблокировка после доступа к мапе

	// Проверка, существует ли уже выражение с таким ID
	if slices.Contains(keys, ID) {
		return nil, rest.NewError("An expression with ID %s is already exists", ID)
	}

	var (
		date  = time.Now()
		value = -1
	)
	if err := errors.New(""); args != nil {
		date = args[0].(time.Time)
		value, err = strconv.Atoi(args[1].(string))
		if err != nil {
			return nil, rest.NewError("Invalid value %s", args[1].(string))
		}
	}

	var ex, err = NewExpression(expr, date, value)
	if err != nil {
		return nil, err
	}

	express.mu.Lock()
	express.IDs[ID] = ex
	express.mu.Unlock()

	if value == -1 {
		go calculator.Calculator(ex) // Запуск вычисления выражения в отдельной горутине
	}

	if err != nil {
		return nil, err
	}

	return ex, nil
}

// Delete удаляет выражения по их ID.
func (express *Expressions) Delete(IDs ...string) {
	defer express.mu.Unlock()
	express.mu.Lock()
	for _, key := range IDs {
		delete(express.IDs, key)
	}

	_ = express.UploadExpressions("data/data_expressions.csv")
}

// Lock блокирует мьютекс для внешнего доступа
func (express *Expressions) Lock() {
	express.mu.Lock()
}

// Unlock разблокирует мьютекс после внешнего доступа
func (express *Expressions) Unlock() {
	express.mu.Unlock()
}

// DownloadExpressions загружает выражения из файла CSV.
func (express *Expressions) DownloadExpressions(name string) error {
	var info, err = data.OpenCSV(name, ';')
	if err != nil {
		return err
	}

	for i, val := range info {
		if i == 0 {
			continue // Пропускаем заголовок файла
		}

		if val[2] != "-1" {
			var expr, err = NewExpression(val[1])
			digit, err := strconv.Atoi(val[2])
			if err != nil {
				return err
			}
			expr.Value = digit

			express.Lock()
			express.IDs[val[0]] = expr
			express.Unlock()
		} else {
			_, err = express.AddExpression(val[0], val[1])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UploadExpressions сохраняет текущие выражения в файл CSV.
func (express *Expressions) UploadExpressions(name string) error {
	var csvFile = make([][]string, 0)
	csvFile = append(csvFile, []string{"ID", "Expression", "Value"})

	for key, val := range express.GetExpressions() {
		var expr = []string{key, val.Express, strconv.Itoa(val.Value)}
		csvFile = append(csvFile, expr)
	}

	var err = data.WriteCSV(csvFile, name, ';')
	return err
}

// GetExpression возвращает выражение по ID, если оно существует, или ошибку.
func (express *Expressions) GetExpression(ID string) (*rest.Expression, error) {
	express.mu.Lock()
	var expr, ok = express.IDs[ID]
	express.mu.Unlock()

	if !ok {
		return nil, rest.NewError("There is no such expression: %d", ID)
	}

	return expr, nil
}

// GetExpressions возвращает копию всех выражений в мапе.
func (express *Expressions) GetExpressions() map[string]*rest.Expression {
	var expressions = map[string]*rest.Expression{}

	express.mu.Lock()
	for key, value := range express.IDs {
		expressions[key] = value
	}
	express.mu.Unlock()

	return expressions
}

// NewExpression создает новый объект Expression с заданным арифметическим выражением.
func NewExpression(express string, args ...interface{}) (*rest.Expression, error) {
	var (
		date          = time.Now()
		value         = -1
		duration, err = calculator.CalculationTime(express)
	)
	if err != nil {
		return nil, err
	}

	if args != nil {
		date = args[0].(time.Time)
		value = args[1].(int)
	}

	return &rest.Expression{
		Value:      value, // Начальное значение, означает отсутствие результата
		Express:    express,
		Result:     make(chan int),
		ErrCh:      make(chan error),
		Created:    date,
		Expiration: duration,
	}, nil
}
