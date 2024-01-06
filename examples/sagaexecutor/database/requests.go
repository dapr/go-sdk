package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/bindings/postgres"
	"github.com/stevef1uk/sagaexecutor/encodedecode"
)

const (
	stateInsert     = "INSERT INTO sagastate (key, value) values ('%s', '%s')"
	stateDelete     = "DELETE FROM sagastate WHERE key = '%s';"
	stateSelect     = "SELECT key, value FROM sagastate;"
	theRowsAffected = "rows-affected"
	operationExec   = "exec"
	sql             = "sql"
)

type StateRecord struct {
	Key   string
	Value string
}

var req = &bindings.InvokeRequest{
	Operation: operationExec,
	Metadata:  map[string]string{},
}

func (r *StateRecord) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		log.Printf("StateRecord error marshalling %v, err = %s\n", r, err)
	}
	return string(data)
}

func GetStateRecords(ctx context.Context, the_db *postgres.Postgres) ([]StateRecord, error) {
	var err error

	req.Operation = "query"
	req.Metadata["sql"] = stateSelect
	res, err := the_db.Invoke(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Error on Query %s", err)
	}

	ret := getTheRows(res.Data)
	return ret, err
}

func Delete(ctx context.Context, the_db *postgres.Postgres, key string) error {
	var err error

	log.Printf("DB:Delete Key = %s\n", key)

	req.Operation = operationExec
	req.Metadata[sql] = fmt.Sprintf(stateDelete, key)
	res, err := the_db.Invoke(ctx, req)
	if err != nil {
		return fmt.Errorf("Error on delete for key %s, err %s", key, err)
	}
	if res.Metadata[theRowsAffected] != "1" {
		return fmt.Errorf("Error on delete row count %s for key = %s", res.Metadata[theRowsAffected], key)
	}
	return err
}

func StoreState(ctx context.Context, the_db *postgres.Postgres, key string, value []byte) error {
	var err error

	log.Printf("DB:Store Key = %s\n", key)

	req.Operation = operationExec
	req.Metadata[sql] = fmt.Sprintf(stateInsert, key, encodedecode.EncodeData(value))
	res, err := the_db.Invoke(ctx, req)
	if err != nil {
		return fmt.Errorf("Error on insert for key %s %s", key, err)
	}
	if res.Metadata[theRowsAffected] != "1" {
		return fmt.Errorf("error on insert row count wrong for key %s", key)
	}

	return err
}

// [[\"one\",\"two\"],[\"mykey\",\"eyJhcHBfaWQiOnRlc3QxLCJzZXJ2aWNlIjp0ZXN0c2VydmljZSwidG9rZW4iOmFiY2QxMjMsImNhbGxiYWNrX3NlcnZpY2UiOmR1bW15LCJwYXJhbXMiOmUzMD0sImV2ZW50IjogdHJ1ZSwidGltZW91dCI6MTAsImxvZ3RpbWUiOjIwMjQtMDEtMDMgMTU6MTI6NDEuOTQ4MDIgKzAwMDAgVVRDfQ==\"]]"
func getTheRows(input []byte) []StateRecord {
	var ret []StateRecord

	var decodedData [][]string

	// Unmarshal the JSON string
	err := json.Unmarshal(input, &decodedData)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil
	}

	ret = make([]StateRecord, len(decodedData))
	for i := 0; i < len(decodedData); i++ {
		ret[i].Key = decodedData[i][0]
		ret[i].Value = decodedData[i][1]
		//fmt.Printf("Row %d from DB = Key = %s, Value = %s\n", i, ret[i].Key, ret[i].Value)
		ret[i].Value = encodedecode.DecodeData(ret[i].Value)
	}

	return ret
}
