package zoomalert

type zoomMessage struct {
	RobotJID  string      `json:"robot_jid"`
	ToJID     string      `json:"to_jid"`
	AccountID string      `json:"account_id"`
	Content   ZoomContent `json:"content"`
}

type ZoomContent struct {
	Head   ZoomHead   `json:"head"`
	Body   []any      `json:"body"` // mixed slice (message, fields, actions)
	Footer ZoomFooter `json:"footer"`
}

type ZoomHead struct {
	Text    string      `json:"text"`
	Style   ZoomStyle   `json:"style"`
	SubHead ZoomSubhead `json:"sub_head"`
}
type ZoomStyle struct {
	Color string `json:"color"`
	Bold  bool   `json:"bold"`
}
type ZoomSubhead struct {
	Text string `json:"text"`
}

type ZoomFooter struct {
	Text string `json:"text"`
}

type Field struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type FieldsBlock struct {
	Type  string  `json:"type"`
	Items []Field `json:"items"`
}
type Action struct {
	Text  string `json:"text"`
	Value string `json:"value"`
	Style string `json:"style"`
}
type ActionsBlock struct {
	Type  string   `json:"type"`
	Items []Action `json:"items"`
}
