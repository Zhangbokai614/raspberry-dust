package upload

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	DB DB `yaml:"database"`
}

type DB struct {
	Host   string `yaml:"host"`
	Name   string `yaml:"name"`
	Passwd string `yaml:"passwd"`
}

func ConnectDatabase() (*sql.DB, error) {
	file, err := ioutil.ReadFile("./database.yaml")
	if err != nil {
		return nil, err
	}

	conf := Config{}

	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return nil, err
	}

	dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/project?parseTime=true", conf.DB.Name, conf.DB.Passwd, conf.DB.Host))
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}
