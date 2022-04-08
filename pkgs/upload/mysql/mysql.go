package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	DBName    = "sensor"
	TableName = "dust"
)

const (
	mysqlCreateUploadDatabase = iota
	mysqlCreateUploadTable
	mysqlInsertUpload
	mysqlQueryUpload
)

var (
	//ErrNoRows -
	ErrNoRows        = errors.New("there is no such data in database")
	errInvalidInsert = errors.New("upload file: insert affected 0 rows")

	SQLString = []string{
		fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s;`, DBName),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
			id		    	BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			query_time  	DATETIME NOT NULL,
			dust			SMALLINT NOT NULL,
			PRIMARY KEY (id)
		)  ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;`, DBName, TableName),
		fmt.Sprintf(`INSERT INTO %s.%s (query_time, dust) VALUES (?, ?)`, DBName, TableName),
		fmt.Sprintf(`SELECT query_time, dust FROM %s.%s WHERE query_time BETWEEN ? AND ?`, DBName, TableName),
	}
)

func CreateDatabase(db *sql.DB) error {
	_, err := db.Exec(SQLString[mysqlCreateUploadDatabase])
	if err != nil {
		return err
	}

	return nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(SQLString[mysqlCreateUploadTable])
	if err != nil {
		return err
	}

	return nil
}

func InsertTable(db *sql.DB, queryTime time.Time, dust int) error {
	result, err := db.Exec(SQLString[mysqlInsertUpload], queryTime, dust)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return errInvalidInsert
	}

	return nil
}

type QueryDust struct {
	QueryTime time.Time
	Dust      int
}

func QueryTable(db *sql.DB, startTime, endTime time.Time) ([]*QueryDust, error) {
	rows, err := db.Query(SQLString[mysqlQueryUpload], startTime, endTime)
	if err != nil {
		return nil, err
	}

	var result []*QueryDust

	for rows.Next() {
		var (
			query_time time.Time
			dust       int
		)

		if err := rows.Scan(&query_time, &dust); err != nil {
			return nil, err
		}

		result = append(result, &QueryDust{
			QueryTime: query_time,
			Dust:      dust,
		})
	}

	return result, nil
}
