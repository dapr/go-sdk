package api

type ActorReminderParams struct {
	Data    []byte `json:"data"`
	DueTime string `json:"dueTime"`
	Period  string `json:"period"`
}
