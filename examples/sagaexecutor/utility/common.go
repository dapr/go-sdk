package utility

import (
	"encoding/base64"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	database "github.com/stevef1uk/sagaexecutor/database"
)

type Start_stop struct {
	App_id           string    `json:"app_id"`
	Service          string    `json:"service"`
	Token            string    `json:"token"`
	Callback_service string    `json:"callback_service"`
	Params           string    `json:"params"`
	Timeout          int       `json:"timeout"`
	Event            bool      `json:"event"`
	LogTime          time.Time `json:"logtime"`
}

type OrderedMessage struct {
	OrderingField string
	Data          []byte
}

const (
	Start  = true
	Stop   = false
	layout = "2006-01-02 150405"
)

// Written to handle input like this. I hope there is an easier way to do this?
// input := `"app_id":sagatxs,"service":serv1,"token":abcdefg1235,"callback_service":localhost,"params":{},"Timeout":100,"TimeLogged":2023-12-16 13:09:05.837307312 +0000 UTC`
func getMapFromString(input string) map[string]string {

	var m map[string]string = make(map[string]string)
	// Remove first characacter as this will be the json { othwerwise it forms part of the first key
	input2 := strings.Replace(input[1:], `"`, ``, -1)
	//fmt.Printf("input2: %s \n", input2)

	split1 := regexp.MustCompile(",").Split(input2, -1)
	for _, v := range split1 {
		split2 := regexp.MustCompile(`:`).Split(v, -1)
		//fmt.Printf("split2: %s \n", split2)
		key := ""
		for i, j := range split2 {
			//fmt.Printf("j: %s \n", j)
			if i == 0 {
				key = j
				//fmt.Printf("key: %s \n", key)
				m[key] = ""
			} else {
				m[key] = m[key] + j
				//fmt.Printf("m[%s] = %s \n", key, m[key])
			}
		}
	}
	//fmt.Printf("map = %v\n", m)
	return m
}

func ProcessRecord(input database.StateRecord, skip_time bool) Start_stop {
	var log_entry Start_stop
	var mymap map[string]string
	var rawDecodedText []byte

	rawDecodedText, err := base64.StdEncoding.DecodeString(input.Value)
	if err != nil {
		log.Printf("Base64 decode failed! %s\n", err)
		panic(err)
	}
	mymap = getMapFromString(string(rawDecodedText))
	if !skip_time {
		time_logtime := mymap["logtime"]
		if time_logtime != "" {
			time_tmp := time_logtime[0:17]
			log.Printf("time_tmp = %s. time_tmp = %s\n", time_logtime, time_tmp)
			log_entry.LogTime, err = time.Parse(layout, time_tmp)
			if err != nil {
				log.Printf("Error parsing time %s\n", err)
			}
			//log.Printf("parsed time = %v\n", log_entry.LogTime)
		}
	}
	log_entry.App_id = mymap["app_id"]
	if mymap["event"] == "true" {
		log_entry.Event = Start
	} else {
		log_entry.Event = Stop
	}
	log.Printf("App_id = %s\n", mymap["app_id"])
	log_entry.Service = mymap["service"]
	log_entry.Token = mymap["token"]
	log_entry.Timeout, _ = strconv.Atoi(mymap["timeout"])
	log_entry.Callback_service = mymap["callback_service"]
	tmp, err := strconv.Unquote(mymap["params"])
	if err != nil {
		tmp = mymap["params"]
	}
	var tmp_b []byte = make([]byte, len(tmp))
	_, _ = base64.StdEncoding.Decode(tmp_b, []byte(mymap["params"]))
	log_entry.Params = string(tmp_b)
	//log.Printf("Log Entry reconstructed = %v\n", log_entry)
	return log_entry
}
