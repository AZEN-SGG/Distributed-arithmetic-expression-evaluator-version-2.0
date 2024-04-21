package calculator

import (
	"Distributed-arithmetic-expression-evaluator-version-2.0/rest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	alphabet           = []int32{40, 41, 42, 47, 43, 45}
	canBe              = []int32{48, 49, 50, 51, 52, 43, 45, 42, 40, 41, 47, 53, 54, 55, 56, 57}
	ArithmeticExecTime = map[int32]time.Duration{43: time.Millisecond * 500, 45: time.Millisecond * 750,
		42: time.Millisecond * 1000, 47: time.Millisecond * 1500}
	ComputingPower []int32
)

type Operation struct {
	operator int32
	value    int64
}

func FormatOperation(opera int32, val string) (*Operation, error) {
	var digit, err = strconv.ParseInt(strings.Replace(val, "-", "", -1), 10, 0)

	if err != nil {
		return nil, err
	}

	return &Operation{
		operator: opera,
		value:    digit,
	}, nil
}

func MathOperation(operations ...*Operation) {
	for _, val := range operations {
		ArithmeticExecTime[val.operator] = time.UnixMilli(val.value).Sub(time.UnixMilli(0))
	}
}

func Waiter(value1, value2 int, operate int32) int {
	time.Sleep(ArithmeticExecTime[operate])

	switch operate {
	case 42:
		return value1 * value2

	case 43:
		return value1 + value2

	case 45:
		return value1 - value2

	case 47:
		return value1 / value2

	default:
		return -1
	}
}

// Delimiter Распределяет подсчёт на разные ярусы
func Delimiter(expr string) ([]int32, []int, error) {
	var queue = make([]int32, 0)
	var values = make([]int, 0)
	var value string

	for _, val := range expr {
		if slices.Contains(alphabet, val) {
			if value != "" {
				num, err := strconv.Atoi(value)

				if err != nil {
					return nil, nil, rest.NewError("Extraneous characters found in expression: %s", value)
				}

				values = append(values, num)
				value = ""
			}

			queue = append(queue, val)
		} else {
			value += string(val)
		}
	}

	if value != "" {
		num, err := strconv.Atoi(value)

		if err != nil {
			return nil, nil, rest.NewError("Extraneous characters found in expression: %s", value)
		}

		values = append(values, num)
	}

	return queue, values, nil
}

// Distributor Разделяет значения на множество действий
func Distributor(queue []int32, values []int) ([][]int32, [][]int, error) {
	var expresses = [][]int32{{}}
	var sortValues = [][]int{{}}

	var indexes = []int{0}
	var indexLastExpression = -1
	var indexValue = 0 // Индекс числа из Value
	var index int

	for i, val := range queue {
		switch val {
		case 40: // (
			indexes = append(indexes, len(expresses))
			sortValues = append(sortValues, make([]int, 0))
			expresses = append(expresses, make([]int32, 0))

		case 41: // )
			index = rest.Last(indexes)

			if indexLastExpression != -1 {
				sortValues[index] = append(sortValues[index], -indexLastExpression)
				indexLastExpression = -1
			} else {
				sortValues[index] = append(sortValues[index], values[indexValue])
				indexValue++
			}

			indexLastExpression = index

			indexes = slices.Delete(indexes, len(indexes)-1, len(indexes))

			if i == len(queue)-1 {
				sortValues[rest.Last(indexes)] = append(sortValues[rest.Last(indexes)], -indexLastExpression)
			}

		default:
			index = rest.Last(indexes)

			if indexLastExpression != -1 {
				sortValues[index] = append(sortValues[index], -indexLastExpression)
				indexLastExpression = -1
			} else {
				sortValues[index] = append(sortValues[index], values[indexValue])
				indexValue++
			}

			expresses[index] = append(expresses[index], val)
		}
	}

	if index == 0 {
		if indexLastExpression != -1 {
			sortValues[index] = append(sortValues[index], -indexLastExpression)
			indexLastExpression = -1
		} else {
			sortValues[index] = append(sortValues[index], values[indexValue])
			indexValue++
		}
	}

	return expresses, sortValues, nil
}

func Mathematician(expresses [][]int32, values [][]int) int {
	var doneCh = make(chan []int)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go Proletarian(&wg, expresses, values, 0, &doneCh)

	go func() {
		defer close(doneCh)
		wg.Wait()
	}()

	for i := range doneCh {
		return i[0]
	}

	return -1
}

func Proletarian(wg *sync.WaitGroup, expresses [][]int32, values [][]int, index int, outCh *chan []int) {
	defer wg.Done()
	var valueWG = sync.WaitGroup{}

	var expr = expresses[index]
	var value = values[index]
	var calculated = make([]int, len(value))

	var routine = ArithmeticSorter(expr)

	var value1, value2 int

	for _, i := range routine {
		ComputingPower = append(ComputingPower, expr[i])

		if calculated[i] == 0 {
			value1 = value[i]
		} else {
			value1 = calculated[i]
		}

		if calculated[i+1] == 0 {
			value2 = value[i+1]
		} else {
			value2 = calculated[i+1]
		}

		var valueCh = make(chan []int)

		if value1 <= 0 {
			valueWG.Add(1)
			go Proletarian(&valueWG, expresses, values, -value1, &valueCh)
		}

		if value2 <= 0 {
			valueWG.Add(1)
			go Proletarian(&valueWG, expresses, values, -value2, &valueCh)
		}

		go func() {
			defer close(valueCh)
			valueWG.Wait()
		}()

		for val := range valueCh {
			if val[1] == -value1 {
				value1 = val[0]
			} else {
				value2 = val[0]
			}
		}

		ComputingPower = slices.Delete(ComputingPower, slices.Index(ComputingPower, expr[i]), slices.Index(ComputingPower, expr[i])+1)
		calculated[i] = Waiter(value1, value2, expr[i])
		calculated[i+1] = calculated[i]
	}

	*outCh <- []int{calculated[rest.Last(routine)], index}
}

// ArithmeticSorter Сортирует операции по важности по убыванию
func ArithmeticSorter(expr []int32) []int {
	var routine []int
	var routineStrong []int

	for i, val := range expr {
		if val == 42 || val == 47 {
			routineStrong = append(routineStrong, i)
		} else if val == 43 || val == 45 {
			routine = append(routine, i)
		}
	}

	return append(routineStrong, routine...)
}

// CalculationTime Считает примерное время выполнения операции
func CalculationTime(expr string) (time.Duration, error) {
	expr, err := PreparingExpression(expr)

	if err != nil {
		return 0, err
	}

	expresses, _, err := Delimiter(expr)

	if err != nil {
		return 0, err
	}

	var workingHours time.Duration

	for _, expr := range expresses {
		workingHours += ArithmeticExecTime[expr]
	}

	return workingHours, nil
}

// PreparingExpression Проверяет выражение на правильность формулировки и форматирует
func PreparingExpression(expr string) (string, error) {
	var parenthesis = 0

	expr = strings.Replace(expr, " ", "", strings.Count(expr, " "))

	var dataReplace = map[string]string{"+-": "-", "--": "+", "++": "+"}

	for i, val := range expr {
		if val == 40 {
			parenthesis++
		} else if val == 41 {
			parenthesis--
		}

		if parenthesis < 0 {
			return "", rest.NewError("Extra closed parenthesis: %s", strconv.Itoa(i))
		}
	}

	if parenthesis != 0 {
		return "", rest.NewError("Extra open parenthesis")
	}

	expr = strings.Trim(expr, "/*+-")

	for key, val := range dataReplace {
		for strings.Count(expr, key) != 0 {
			expr = strings.Replace(expr, key, val, strings.Count(expr, key))
		}
	}

	for i, elem := range expr {
		if !slices.Contains(canBe, elem) {
			return "", rest.NewError("Foreign character detected: %s", string(elem))
		} else if slices.Contains(alphabet, elem) && elem > 41 && ((slices.Contains(alphabet, int32(expr[i-1])) && expr[i-1] > 41) || (slices.Contains(alphabet, int32(expr[i+1])) && expr[i+1] > 41)) {
			return "", rest.NewError("Incorrect expression: %s", expr[i-1:i+2])
		}
	}

	return expr, nil
}

// Calculator Решает арифметическое выражение
func Calculator(express *rest.Expression) {
	defer express.Close()

	expr, err := PreparingExpression(express.Express)

	if err != nil {
		express.ErrCh <- err
		return
	}

	queue, values, err := Delimiter(expr)

	if err != nil {
		express.ErrCh <- err
		return
	}

	expresses, value, err := Distributor(queue, values)

	if err != nil {
		express.ErrCh <- err
		return
	}

	for i, val := range expresses {
		if len(val)+1 != len(value[i]) {
			express.ErrCh <- rest.NewError("Incorrect expression")
			return
		} else if len(val) == 0 || len(value[i]) < 2 {
			express.ErrCh <- rest.NewError("Too few arguments")
			return
		}
	}

	answer := Mathematician(expresses, value)

	express.Result <- answer

	return
}
