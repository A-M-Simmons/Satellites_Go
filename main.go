package main

import _ "github.com/go-sql-driver/mysql"
import "database/sql"
import "log"

// TLE Structure
type TLE struct {
	ncid     int
	date     int
	epoch    string
	tleline1 string
	tleline2 string
}

func parseTLEResults(rows *sql.Rows, err error) map[int]TLE {
	var tle TLE
	results := make(map[int]TLE)
	for rows.Next() {
		err = rows.Scan(&tle.ncid, &tle.date, &tle.epoch, &tle.tleline1, &tle.tleline2)
		if err != nil {
			panic(err)
		}
		results[tle.ncid] = tle
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return results
}

func main() {
	// Initialise Variables
	rdsCredJSON := "cred.json"
	reqFile := "request.json"

	// Connect to Database
	db := connectToDB(rdsCredJSON)
	rows, err := db.Query("SELECT NORAD_CAT_ID, DATE, EPOCH, TLE_LINE1, TLE_LINE2 FROM TLE WHERE DATE = 20181213")
	defer db.Close()
	log.Printf("Pulled request from Database")

	// Parse results
	results := parseTLEResults(rows, err)
	log.Printf("Parsed results")

	// Find Satellites
	nearbySatelliteList := findNearbySatellites(results, reqFile)
	matchedMap := getMatchedSatellitesData(nearbySatelliteList, results, reqFile)
	log.Printf("Calculations completed")

	prettyDisplay(matchedMap)
}
