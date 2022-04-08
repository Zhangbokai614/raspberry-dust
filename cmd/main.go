package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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
	dustFilePath = fmt.Sprintf("./data/dust-%s.csv", time.Now().Format("2006-01-02"))
)

func main() {
	// db, err := pangolin.OpenDB(pangolin.DefaultOption(uuid.NewString()))
	// if err != nil {
	// 	log.Fatal(err)
	// }

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

		// if err := db.Insert(time.Now().UnixMicro(), result); err != nil {
		// 	log.Fatal(err)
		// }

		fmt.Printf("😀 Read and save succeed:%s\n", record)
		time.Sleep(frequency)
	}
}
