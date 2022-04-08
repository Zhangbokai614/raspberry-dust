package upload

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"dust/pkgs/upload/mysql"

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

type UploadController struct {
	db *sql.DB
}

func New() (*UploadController, error) {
	file, err := ioutil.ReadFile("./database.yaml")
	if err != nil {
		return nil, err
	}

	conf := Config{}

	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return nil, err
	}

	dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true&loc=Local", conf.DB.Name, conf.DB.Passwd, conf.DB.Host))
	if err != nil {
		return nil, err
	}

	return &UploadController{
		db: dbConn,
	}, nil
}

func (c *UploadController) DBInit() error {
	err := mysql.CreateDatabase(c.db)
	if err != nil {
		return err
	}

	err = mysql.CreateTable(c.db)
	if err != nil {
		return err
	}

	return nil
}

func (c *UploadController) Insert(queryTime time.Time, dust int) error {
	err := mysql.InsertTable(c.db, queryTime, dust)
	if err != nil {
		return err
	}

	return nil
}

func (c *UploadController) Query(w http.ResponseWriter, r *http.Request) {
	parameters := r.URL.Query()
	startTime := parameters.Get("startTime")
	endTime := parameters.Get("endTime")

	loc, err := time.LoadLocation("Local")
	if err != nil {
		w.WriteHeader(500)
	}

	longForm := "2006-01-02"
	st, err := time.ParseInLocation(longForm, startTime, loc)
	if err != nil {
		w.WriteHeader(400)
	}

	et, err := time.ParseInLocation(longForm, endTime, loc)
	if err != nil {
		w.WriteHeader(400)
	}

	result, err := mysql.QueryTable(c.db, st, et)
	if err != nil {
		w.WriteHeader(500)
	}

	msg, _ := json.Marshal(result)

	w.Header().Set("content-type", "text/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Write(msg)
}
