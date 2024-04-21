package rest

import (
	"fmt"
	"time"
)

// Пришлось поместить сюда, чтобы не происходил cycle, так как пакету calculator нужен данный класс

// Expression представляет выражение с его свойствами.
type Expression struct {
	Value      int           // Используется для хранения результата выражения
	Express    string        // Строковое представление выражения, например "2+2"
	Result     chan int      // Канал для получения результата вычисления выражения
	ErrCh      chan error    // Канал для передачи ошибок при вычислении
	Created    time.Time     // Время создания экземпляра выражения
	Expiration time.Duration // Продолжительность жизни выражения
}

// Close метод закрывает каналы ErrCh и Result для освобождения ресурсов.
func (express *Expression) Close() {
	close(express.ErrCh)
	close(express.Result)
}

// GetValue пытается получить значение из канала Result или ошибку из канала ErrCh.
// Возвращает 0 и ошибку, если есть ошибка, или результат, если нет ошибки.
// Если нет доступных значений в каналах, возвращает -1, что означает отсутствие данных.
func (express *Expression) GetValue() (int, error) {
	select {
	case err := <-express.ErrCh: // Чтение из канала ошибок
		return 0, err
	case answer := <-express.Result: // Чтение результата вычисления
		if express.Value == -1 {
			express.Value = answer // Сохранение результата в свойство Value, если оно было инициализировано как -1
		}
		return answer, nil
	default:
		return -1, nil // Возврат -1, если ни одно значение не готово для чтения
	}
}

// Last возвращает последний элемент из слайса любого типа.
func Last[E any](s []E) E {
	return s[len(s)-1]
}

// NewError создает новый экземпляр ошибки с форматированным сообщением.
func NewError(format string, values ...interface{}) error {
	return fmt.Errorf(format, values...)
}

// MapGetKeys возвращает слайс ключей из переданной карты.
// Использует дженерик типы для работы с картами различных типов ключей и значений.
func MapGetKeys[K comparable, V any](m map[K]V) []K {
	var keys = make([]K, len(m))
	var index int
	for key := range m {
		keys[index] = key
		index++
	}
	return keys
}
