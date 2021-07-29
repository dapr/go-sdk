package reminder

type ActorReminderParams struct {
	//MinTimePeriod string `json:"min_time_period"`
	Data    []byte `json:"data"`
	DueTime string `json:"dueTime"`
	Period  string `json:"period"`
}
