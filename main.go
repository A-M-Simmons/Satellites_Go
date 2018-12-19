package main

import _ "github.com/go-sql-driver/mysql"
import "github.com/joshuaferrara/go-satellite"
import "database/sql"
import "time"
import "log"
import "fmt"
import "math"
import "os"
import "encoding/json"
import "io/ioutil"

// TLE Structure
type TLE struct {
	ncid     int
	date     int
	epoch    string
	tleline1 string
	tleline2 string
}

// RDSCreds structure
type RDSCreds struct {
	Name     string `json:"username"`
	Password string `json:"password"`
	Endpoint string `json:"endpoint"`
	Port     string `json:"port"`
	DbName   string `json:"dbname"`
}

type request struct {
	Time     requestTime     `json:"Time"`
	Location requestLocation `json:"Location"`
}

type requestTime struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type requestLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
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

func parseTLEResults(rows *sql.Rows, err error) (results []TLE) {
	timeline := TLE{}
	for rows.Next() {
		err = rows.Scan(&timeline.ncid, &timeline.date, &timeline.epoch, &timeline.tleline1, &timeline.tleline2)
		if err != nil {
			panic(err)
		}
		results = append(results, timeline)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return results
}

func compareLookAngle(elevation float64, position satellite.Vector3, obsCoord satellite.LatLong, t time.Time) (bool, float64) {
	lAngles := satellite.ECIToLookAngles(position, obsCoord, 27.0/10000.0, satellite.JDay(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()))
	if lAngles.El*180.0/math.Pi > elevation {
		return true, 0.0
	}
	return false, elevation - lAngles.El*180.0/math.Pi
}

func addTime(t time.Time, scale float64) time.Time {
	if scale > 90 {
		t = t.Add(time.Second * 120)
	} else if scale > 60 {
		t = t.Add(time.Second * 60)
	} else if scale > 30 {
		t = t.Add(time.Second * 30)
	} else if scale > 10 {
		t = t.Add(time.Second * 15)
	} else {
		t = t.Add(time.Second * 1)
	}
	return t
}

func loadRequestJSON(file string) (request request) {
	jsonFile, err1 := os.Open(file)
	if err1 != nil {
		fmt.Println(err1)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &request)
	return request
}

func main() {
	// Initialise Variables
	rdsCredJSON := "cred.json"
	req := loadRequestJSON("request.json")
	startTime, err := time.Parse("2006-01-02 15:04:05", req.Time.StartTime)
	endTime, err := time.Parse("2006-01-02 15:04:05", req.Time.EndTime)
	var obsCoord satellite.LatLong
	obsCoord.Latitude = req.Location.Latitude * math.Pi / 180.0
	obsCoord.Longitude = req.Location.Longitude * math.Pi / 180.0

	// Connect to Database
	db := connectToDB(rdsCredJSON)
	rows, err := db.Query("SELECT NORAD_CAT_ID, DATE, EPOCH, TLE_LINE1, TLE_LINE2 FROM TLE WHERE DATE = 20181213")
	defer db.Close()
	log.Printf("Pulled request from Database")

	// Parse results
	results := parseTLEResults(rows, err)
	log.Printf("Parsed results")

	// Iterate through records
	for _, record := range results {
		sat := satellite.TLEToSat(record.tleline1, record.tleline2, "wgs72")
		t := startTime
		// Iterate over time
		for ok := true; ok; ok = t.Before(endTime) {
			position, _ := satellite.Propagate(sat, t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
			withinRange, scale := compareLookAngle(85.0, position, obsCoord, t)
			if withinRange {
				break
			}
			t = addTime(t, scale)
		}
	}
	log.Printf("Calculations completed")
}
