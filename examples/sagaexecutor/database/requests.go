package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sleep_time = 250 * time.Millisecond
	max_reties = 4
)

type StateRecord struct {
	Key   string
	Value string
}

func GetStateRecords(ctx context.Context, the_db *pgxpool.Pool) ([]StateRecord, error) {
	var err error
	var ret []StateRecord = make([]StateRecord, 0)

	retries := 0

	//log.Printf("Entered GetStateRecords\n")
	for retries < max_reties {
		//log.Printf("Retries = %v\n", retries)
		rows, err := the_db.Query(ctx, "SELECT key, value FROM sagastate;")
		if err != nil {
			//log.Printf("`Error path` = %v\n", err)
			if err.Error() != "conn busy" { // Should be ErrConnBusy
				return nil, fmt.Errorf("Error: Select query failed: %v", err)
			} else {
				log.Printf("DB Busy for select of state record: %v\n", err)
				time.Sleep(sleep_time)
				retries = retries + 1
			}
		} else {
			// Loop through rows, using Scan to assign column data to struct fields.
			for rows.Next() {
				//log.Printf("Scanning a row \n")
				var state = &StateRecord{}
				if err := rows.Scan(&state.Key, &state.Value); err != nil {
					return nil, fmt.Errorf("Error state query Scan: %v", err)
				}
				//log.Printf("Appending a row key = %s\n", state.Key)
				ret = append(ret, *state)
			}
			defer rows.Close()
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("Error state processing of records: %v", err)
			}
			break
		}
	}

	if retries == max_reties {
		return nil, fmt.Errorf("DB remained busy for too long on select: %s\n", err)
	}

	return ret, nil
}

func Delete(ctx context.Context, the_db *pgxpool.Pool, key string) error {
	var err error

	retries := 0

	log.Printf("DB:Delete Key = %s\n", key)
	tx, err := the_db.Begin(ctx)
	for retries < max_reties {
		res, err := the_db.Exec(ctx, "DELETE FROM sagastate WHERE key = $1;", key)
		if err != nil {
			if err.Error() != "conn busy" { // Should be ErrConnBusy
				return fmt.Errorf("`Delete failed for state record with key %s: %v", key, err)
			} else {
				log.Printf("DB Busy for delete of state record with key %s: %v\n", key, err)
				time.Sleep(sleep_time)
				retries = retries + 1
			}
		} else {
			rowsAffected := res.RowsAffected()

			if rowsAffected > 1 {
				log.Printf("Wrong number of records deleted %v \n", rowsAffected)
			}
			err = tx.Commit(ctx)
			if err != nil {
				log.Printf("DB Commit failed %s\n", err)
			}
			break
		}
	}
	if retries == max_reties {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("DB remained busy for too long on delete for key = %s: err %s \n", key, err)
	}

	return err
}

func StoreState(ctx context.Context, the_db *pgxpool.Pool, key string, value []byte) error {
	var err error

	params := base64.StdEncoding.EncodeToString([]byte(value))
	str := string(params)

	retries := 0
	log.Printf("DB:Store Key = %s\n", key)
	tx, err := the_db.Begin(ctx)
	for retries < max_reties {
		res, err := the_db.Exec(ctx, `INSERT INTO sagastate (key, value) values ($1, $2)`, &key, &str)
		if err != nil {
			if err.Error() != "conn busy" { // Should be ErrConnBusy
				return fmt.Errorf("`Insert failed for state record with key %s: %v", key, err)
			} else {
				log.Printf("DB Busy for insert of state record with key %s: %v\n", key, err)
				time.Sleep(sleep_time)
				retries = retries + 1
			}
		} else {
			rowsAffected := res.RowsAffected()

			if rowsAffected != 1 {
				log.Printf("Wrong number of records for insert %v \n", rowsAffected)
			} else {
				err = tx.Commit(ctx)
				if err != nil {
					log.Printf("DB Commit failed %s\n", err)
				}
			}
			break
		}
	}

	if retries == max_reties {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("DB remained busy for too long on insert for key %s: err %s \n", key, err)
	}

	return err
}
