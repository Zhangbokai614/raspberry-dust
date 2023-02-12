package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"dust/pkgs/chart"
	dust_sensor "dust/pkgs/device"
	upload "dust/pkgs/upload/controller"
	save "dust/pkgs/util"

	"github.com/dovics/pangolin"
	"github.com/dovics/pangolin/lsmt"
)

var (
	//sensor config
	config = &dust_sensor.Config{
		PortName:        "/dev/serial0",
		BaudRate:        10000, // 9600
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 8,
	}

	//query frequency
	frequency = (60 * 5 * time.Second)

	//save csv path
	dustFilePath = fmt.Sprintf("./data/dust-%s.csv", time.Now().Format("2006-01-02"))
)

func main() {
	option := pangolin.DefaultOption("dc7d60be-e9a9-45bb-9f0a-a5c3dcb42e52")
	lsmt.DefaultOption.MemtableSize = 1024
	option.EngineOption = lsmt.DefaultOption

	db, err := pangolin.OpenDB(option)
	if err != nil {
		log.Fatal(err)
	}

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

	go func() {
		http.HandleFunc("/queryDust", dbcon.Query)
		http.HandleFunc("/chart", chart.BarChart)
		http.ListenAndServe(":8081", nil)
	}()

	defer sensorConn.Port.Close()

	for {
		err = sensorConn.SetDeviceMod()
		if err != nil {
			fmt.Printf("Set device mod: %v\nReset\n", err)

			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	for {
		result, err := sensorConn.QueryDust()
		if err != nil {
			fmt.Printf("Query fail: %s\nRequery\n", err)

			for {
				err = sensorConn.SetDeviceMod()
				if err != nil {
					fmt.Printf("Set device mod: %v\nReset\n", err)
					continue
				}
				break
			}

			continue
		}

		record := [][]string{{time.Now().Format("2006-01-02 15:04"), strconv.Itoa(result)}}

		dustFilePath = fmt.Sprintf("./data/dust-%s.csv", time.Now().Format("2006-01-02"))
		err = save.SaveToCsv(record, dustFilePath, false)
		if err != nil {
			log.Fatalf("Save to CSV fail: %s", err)
		}

		if err := dbcon.Insert(time.Now(), result); err != nil {
			log.Fatalf("Upload to database fail: %s", err)
		}

		if err := db.Insert(time.Now().Unix(), result); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("😀 Read and save succeed:%s\n", record)
		time.Sleep(frequency)
	}
}
