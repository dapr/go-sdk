package api

type Metadata struct {
	DBName         string `json:"db_name"`
	BackupLocation string `json:"backup_location"`
}

type DBBackup struct {
	Task     string   `json:"task"`
	Metadata Metadata `json:"metadata"`
}
