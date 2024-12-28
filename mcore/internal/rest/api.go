package rest

import (
	"encoding/json"
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"log"
	"net/http"
)

type Api struct {
	rootPath string
	taskMng  taskMnger
	router   router
	debug    bool
	secrets  []string
	port     string
}

type taskMnger interface {
	GetTask(taskType schema.TaskType) (schema.Task, error)
	ReportTask(taskId int64, status schema.TaskStatus, textMsg string) error
}

type router interface {
	Build(m schema.Message) schema.TaskMsg
}

func NewApi(rootPath string, taskMng taskMnger, router router, debug bool, secrets []string, port string) *Api {
	return &Api{
		rootPath: rootPath,
		taskMng:  taskMng,
		router:   router,
		debug:    debug,
		secrets:  secrets,
		port:     port,
	}
}

type ErrorRes struct {
	Error  string `json:"error"`
	Status string `json:"status"`
}

func (a *Api) Run() error {
	mux := http.NewServeMux()

	getTaskLink := fmt.Sprintf("%s/get-task/", a.rootPath)
	reportTaskLink := fmt.Sprintf("%s/report-task/", a.rootPath)
	addMsgLink := fmt.Sprintf("%s/add-msg/", a.rootPath)

	mux.HandleFunc("POST "+getTaskLink, a.HandlerGetTask)
	mux.HandleFunc("POST "+reportTaskLink, a.HandlerReportTask)
	mux.HandleFunc("POST "+addMsgLink, a.HandlerAddMsg)
	mux.HandleFunc("GET /health/", a.HandlerHealth)

	var h http.Handler
	h = mux
	if a.debug {
		log.Println("debug is on")
		h = middleLog(h)
	}
	h = middleAuth(h, a.secrets)

	log.Println("start server port:" + a.port)
	return http.ListenAndServe(":"+a.port, h)
}

func (a *Api) HandlerGetTask(w http.ResponseWriter, req *http.Request) {
	var taskReq schema.GetTaskReq
	err := json.NewDecoder(req.Body).Decode(&taskReq)
	if err != nil {
		getErrResp(w, fmt.Errorf("body GetTask decode err: %w", err))
		return
	}

	task, err := a.taskMng.GetTask(taskReq.TaskType)
	if err != nil {
		getErrResp(w, fmt.Errorf("getTask err: %w", err))
		return
	}

	status := "OK"
	if task.Id == 0 {
		status = "no tasks"
	}

	b, err := json.Marshal(schema.GetTaskRes{
		Data:   task,
		Status: status,
	})
	if err != nil {
		getErrResp(w, fmt.Errorf("response GetTask decode err: %w", err))
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Fatalf("HandlerGetTask can not write answer %s", err.Error())
	}

}

func (a *Api) HandlerReportTask(w http.ResponseWriter, req *http.Request) {
	var rt schema.ReportTaskReq
	err := json.NewDecoder(req.Body).Decode(&rt)
	if err != nil {
		getErrResp(w, fmt.Errorf("body ReportTask decode err: %w", err))
		return
	}

	err = a.taskMng.ReportTask(rt.TaskId, rt.Status, rt.TextMsg)
	if err != nil {
		getErrResp(w, fmt.Errorf("reportTask err: %w", err))
		return
	}

	b, err := json.Marshal(schema.Req{
		Status: "OK",
	})
	if err != nil {
		getErrResp(w, fmt.Errorf("response reportTask decode err: %w", err))
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Fatalf("HandlerReportTask can not write answer %s", err.Error())
	}

}

func (a *Api) HandlerAddMsg(w http.ResponseWriter, req *http.Request) {
	var m schema.Message

	err := json.NewDecoder(req.Body).Decode(&m)
	if err != nil {
		getErrResp(w, fmt.Errorf("body addMsg decode err: %w", err))
		return
	}

	if m.ChatId == 0 || m.UserName == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	t := a.router.Build(m)

	b, err := json.Marshal(schema.AddMsgReq{
		Data:   t,
		Status: "OK",
	})
	if err != nil {
		getErrResp(w, fmt.Errorf("response addMsg decode err: %w", err))
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Fatalf("addMsg can not write answer %s", err.Error())
	}
}

func getErrResp(w http.ResponseWriter, err error) {
	log.Println("handler: " + err.Error())
	b, err := json.Marshal(ErrorRes{
		Error:  err.Error(),
		Status: "error",
	})
	if err != nil {
		log.Fatalf("http handler can not marshal an error %s", err.Error())
	}
	_, err = w.Write(b)
	if err != nil {
		log.Fatalf("http handler cannot return an error %s", err.Error())
	}
	return
}

func middleAuth(next http.Handler, secrets []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		if url.Path == "/health/" || url.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		secret := r.Header.Get("secret")
		for _, s := range secrets {
			if s == secret {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func middleLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Printf("req method: %s, path: %s", r.Method, r.URL.EscapedPath())

		next.ServeHTTP(w, r)
	})
}

func (a *Api) HandlerHealth(w http.ResponseWriter, req *http.Request) {
	type PingRes struct {
		Status string `json:"status"`
	}

	js, _ := json.Marshal(PingRes{Status: "OK"})
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(js)
	if err != nil {
		log.Fatalf("health can not write answer %s", err.Error())
	}
}
