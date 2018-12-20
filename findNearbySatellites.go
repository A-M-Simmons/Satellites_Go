package main

import "github.com/joshuaferrara/go-satellite"
import "fmt"
import "math"
import "os"
import "encoding/json"
import "io/ioutil"
import "time"

type matchedSatelliteContainer struct {
	record []matchedSatelliteRecord
}
type matchedSatelliteRecord struct {
	time    time.Time
	entered bool
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
	Elevation float64 `json:"elevation"`
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

func compareLookAngle(elevation float64, position satellite.Vector3, obsCoord satellite.LatLong, t time.Time) (bool, float64) {
	lAngles := satellite.ECIToLookAngles(position, obsCoord, 27.0/10000.0, satellite.JDay(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()))
	if lAngles.El*180.0/math.Pi > elevation {
		return true, 0.0
	}
	return false, elevation - lAngles.El*180.0/math.Pi
}

// Function that gets time stamps of when satellite is in elevation
func getMatchedSatellitesData(matched []int, results map[int]TLE, reqFile string) map[int]matchedSatelliteContainer {
	req := loadRequestJSON(reqFile)
	startTime, _ := time.Parse("2006-01-02 15:04:05", req.Time.StartTime)
	endTime, _ := time.Parse("2006-01-02 15:04:05", req.Time.EndTime)
	var obsCoord satellite.LatLong
	obsCoord.Latitude = req.Location.Latitude * math.Pi / 180.0
	obsCoord.Longitude = req.Location.Longitude * math.Pi / 180.0
	wasMatched := false
	matchedMap := make(map[int]matchedSatelliteContainer)
	for _, NCID := range matched {
		sat := satellite.TLEToSat(results[NCID].tleline1, results[NCID].tleline2, "wgs72")
		t := startTime
		var satelliteMatchedRecord []matchedSatelliteRecord
		// Iterate over time
		for ok := true; ok; ok = t.Before(endTime) {
			position, _ := satellite.Propagate(sat, t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
			withinRange, _ := compareLookAngle(req.Location.Elevation, position, obsCoord, t)
			if (withinRange && !wasMatched) || (!withinRange && wasMatched) {
				entry := matchedSatelliteRecord{time: t, entered: !wasMatched}
				satelliteMatchedRecord = append(satelliteMatchedRecord, entry)
				wasMatched = !wasMatched
			}
			t = t.Add(time.Second * 1)
		}
		matchedMap[NCID] = matchedSatelliteContainer{record: satelliteMatchedRecord}
	}
	return matchedMap
}

func findNearbySatellites(results map[int]TLE, reqFile string) (matched []int) {
	req := loadRequestJSON(reqFile)
	startTime, _ := time.Parse("2006-01-02 15:04:05", req.Time.StartTime)
	endTime, _ := time.Parse("2006-01-02 15:04:05", req.Time.EndTime)
	var obsCoord satellite.LatLong
	obsCoord.Latitude = req.Location.Latitude * math.Pi / 180.0
	obsCoord.Longitude = req.Location.Longitude * math.Pi / 180.0

	// Iterate through records
	for _, record := range results {
		sat := satellite.TLEToSat(record.tleline1, record.tleline2, "wgs72")
		t := startTime
		// Iterate over time
		for ok := true; ok; ok = t.Before(endTime) {
			position, _ := satellite.Propagate(sat, t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
			withinRange, scale := compareLookAngle(req.Location.Elevation, position, obsCoord, t)
			if withinRange {
				matched = append(matched, record.ncid)
				break
			}
			t = addTime(t, scale)
		}
	}
	return matched
}
