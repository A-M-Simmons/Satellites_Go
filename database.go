package main

import _ "github.com/go-sql-driver/mysql"
import "database/sql"
import "os"
import "fmt"
import "encoding/json"
import "io/ioutil"

// RDSCreds structure
type RDSCreds struct {
	Name     string `json:"username"`
	Password string `json:"password"`
	Endpoint string `json:"endpoint"`
	Port     string `json:"port"`
	DbName   string `json:"dbname"`
}

func connectToDB(file string) (db *sql.DB) {
	jsonFile, err1 := os.Open(file)
	if err1 != nil {
		fmt.Println(err1)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var creds RDSCreds
	json.Unmarshal(byteValue, &creds)
	db, err2 := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", creds.Name, creds.Password, creds.Endpoint, creds.Port, creds.DbName))
	if err2 != nil {
		panic(err2.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	return db
}
