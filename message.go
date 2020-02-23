package signaling

type Message struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}
