package api

type ActorTimerParam struct {
	CallBack string `json:"callback"`
	Data     []byte `json:"data"`
	DueTime  string `json:"dueTime"`
	Period   string `json:"period"`
	TTL      string `json:"ttl"`
}
