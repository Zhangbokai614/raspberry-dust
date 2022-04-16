package chart

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-echarts/go-echarts/charts"
)

func BarChart(w http.ResponseWriter, _ *http.Request) {
	dustFilePath := fmt.Sprintf("./data/dust-%s.csv", time.Now().Format("2006-01-02"))

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.ToolboxOpts{Show: true})

	_, err := os.Lstat(dustFilePath)
	if err != nil && os.IsNotExist(err) {
		time.Sleep(10 * time.Second)
	}

	f, err := os.Open(dustFilePath)
	if err != nil {
		log.Println(err)
	}

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Println(err)
	}

	f.Close()

	var nameItems = []string{}
	var valueItems = []string{}

	for _, item := range records {
		nameItems = append(nameItems, item[0])
		valueItems = append(valueItems, item[1])
	}

	bar.AddXAxis(nameItems).
		AddYAxis("PM2.5", valueItems)

	c, err := os.Create("bar.html")
	if err != nil {
		log.Println(err)
	}
	bar.Render(w, c)
}
