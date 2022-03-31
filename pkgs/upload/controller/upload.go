package upload

import (
	"database/sql"
	"fmt"
	"io/ioutil"

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

	dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true", conf.DB.Name, conf.DB.Passwd, conf.DB.Host))
	if err != nil {
		return nil, err
	}

	return &UploadController{
		db: dbConn,
	}, nil
}

func (c *UploadController) DBInit() error {
	println("a")

	err := mysql.CreateDatabase(c.db)
	if err != nil {
		return err
	}
	println("b")

	err = mysql.CreateTable(c.db)
	if err != nil {
		return err
	}
	println("c")

	return nil
}

func (c *UploadController) Upload(dust int) error {
	err := mysql.InsertTable(c.db, dust)
	if err != nil {
		return err
	}

	return nil
}
