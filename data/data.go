package data

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"
)

func OpenCSV(name string, comma rune) ([][]string, error) {
	var file, err = os.Open(name)

	if err != nil {
		return nil, err
	}

	var reader = csv.NewReader(file)
	reader.Comma = comma
	var data [][]string

	data, err = reader.ReadAll()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func WriteCSV(data [][]string, name string, comma rune) error {
	var file, err = os.Create(name)

	if err != nil {
		return err
	}

	var writer = csv.NewWriter(file)

	writer.Comma = comma
	err = writer.WriteAll(data)

	if err != nil {
		return err
	}

	return nil
}

func DownloadArithmetic(arithmetic map[int32]time.Duration, name string) error {
	var info, err = OpenCSV(name, ';')

	if err != nil {
		return err
	}

	for i, val := range info {
		if i == 0 {
			continue
		}

		var digit, err = strconv.ParseInt(strings.Replace(val[1], "-", "", -1), 10, 0)

		if err != nil {
			return err
		}

		var symbol, _ = strconv.Atoi(val[0])
		arithmetic[int32(symbol)] = time.UnixMilli(digit).Sub(time.UnixMilli(0))
	}

	return nil
}

func UploadArithmetic(arithmetic map[int32]time.Duration, name string) error {
	var info = make([][]string, 0)
	info = append(info, []string{"element", "value"})

	for key, val := range arithmetic {
		var mills = strconv.Itoa(int(val.Milliseconds()))

		info = append(info, []string{strconv.Itoa(int(key)), mills})
	}

	var err = WriteCSV(info, name, ';')
	return err
}
