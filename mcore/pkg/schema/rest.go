package schema

type Req struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type GetTaskReq struct {
	TaskType TaskType `json:"taskType"`
}

type GetTaskRes struct {
	Data   Task   `json:"data"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

type ReportTaskReq struct {
	TaskId  int64      `json:"taskId"`
	Status  TaskStatus `json:"status"`
	TextMsg string     `json:"textMsg"`
}

type AddMsgReq struct {
	Data   TaskMsg `json:"taskMsg"`
	Status string  `json:"status"`
	Error  string  `json:"error"`
}
