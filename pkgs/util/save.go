package save

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

func SaveToCsv(result int, path string, tableHeader bool) (err error) {
	fmt.Println("3")

	record := [][]string{{time.Now().Format("2006-01-02 15:04"), strconv.Itoa(result)}}

	_, err = os.Lstat(path)
	if err != nil && os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if tableHeader {
			w := csv.NewWriter(file)
			if err = w.WriteAll([][]string{{"date", "PM2.5"}}); err != nil {
				return err
			}
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	w := csv.NewWriter(file)
	if err = w.WriteAll(record); err != nil {
		return err
	}

	return nil
}
