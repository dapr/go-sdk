package utility

import (
	"encoding/json"
	"log"
	"time"

	"github.com/stevef1uk/sagaexecutor/database"
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
	Start            = true
	Stop             = false
	layout           = "2006-01-02 15:04:05"
	ExpiryDateLayout = layout
)

func ProcessRecord(theInput database.StateRecord, skip_time bool) Start_stop {
	log_entry := &Start_stop{}
	//var mymap map[string]string

	//log.Printf("ProcessRecord Data In = :%v\n", theInput)

	err := json.Unmarshal([]byte(theInput.Value), &log_entry)
	if err != nil {
		log.Printf("Unmarshall in ProcessRecord failed! %s\n", err)
	}

	/*var tmp_b []byte = make([]byte, len(log_entry.Params))
	_, _ = base64.StdEncoding.Decode(tmp_b, []byte(log_entry.Params))
	log_entry.Params = string(tmp_b)*/
	//log.Printf("Log Entry reconstructed = %v\n", log_entry)
	return *log_entry
}
