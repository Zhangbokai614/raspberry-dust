package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/charts"

	dust_sensor "dust/pkgs/device"
	upload "dust/pkgs/upload/controller"
	save "dust/pkgs/util"
)

var (
	//sensor config
	config = &dust_sensor.Config{
		PortName:        "/dev/serial0",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 8,
	}

	//query frequency
	frequency = (60 * 5 * time.Second)

	//save csv path
	dustFilePath = "./data/2-2122-dust.csv"
)

func main() {
	sensorConn, err := dust_sensor.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	dbcon, err := upload.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := dbcon.DBInit(); err != nil {
		log.Fatal(err)
	}

	defer sensorConn.Port.Close()

	for {
		err = sensorConn.SetDeviceMod()
		if err != nil {
			fmt.Printf("Set device mod: %v\nReset\n", err)
			continue
		}
		break
	}

	go func(dbcon *upload.UploadController) {
		for {
			fmt.Println("5")
			result, err := sensorConn.QueryDust()
			if err != nil {
				fmt.Printf("Query fail: %s\nRequery\n", err)
				continue
			}

			fmt.Println("6")
			record := [][]string{{time.Now().Format("2006-01-02 15:04"), strconv.Itoa(result)}}

			err = save.SaveToCsv(record, dustFilePath, false)
			if err != nil {
				log.Fatalf("Save to CSV fail: %s", err)
			}

			if err := dbcon.Upload(result); err != nil {
				log.Fatalf("Upload to database fail: %s", err)
			}

			fmt.Println("7")
			fmt.Printf("😀 Read and save succeed:%s\n", record)
			time.Sleep(frequency)
		}
	}(dbcon)

	http.HandleFunc("/dust", chart)
	http.ListenAndServe(":8081", nil)
}

func chart(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("4")

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

	len := len(records)
	if len >= 60*24 {
		for _, item := range records[:(60 * 24)] {
			nameItems = append(nameItems, item[0])
			valueItems = append(valueItems, item[1])
		}
	}

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
