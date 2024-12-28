package mcoreclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	addr    string
	secret  string
	timeout time.Duration
}

const (
	addMsgUrl     = "/add-msg/"
	getTaskUrl    = "/get-task/"
	reportTaskUrl = "/report-task/"
)

func NewClient(addr, secret string) *Client {
	return &Client{
		addr:    addr,
		secret:  secret,
		timeout: 10 * time.Second,
	}
}

func (c *Client) AddMsg(msgRes schema.Message) (schema.AddMsgReq, error) {
	var mr schema.AddMsgReq
	body, err := json.Marshal(msgRes)
	if err != nil {
		return mr, fmt.Errorf("addmsg marshal err %w", err)
	}

	reqBody, err := c.doPost(addMsgUrl, body)
	if err != nil {
		return mr, fmt.Errorf("addmsg doPost: %w", err)
	}
	err = json.Unmarshal(reqBody, &mr)
	if err != nil {
		return mr, fmt.Errorf("addmsg Unmarshal req: %w", err)
	}

	return mr, nil
}

func (c *Client) GetTask(taskReq schema.GetTaskReq) (schema.GetTaskRes, error) {
	var tr schema.GetTaskRes
	body, err := json.Marshal(taskReq)
	if err != nil {
		return tr, fmt.Errorf("getTask marshal err %w", err)
	}

	reqBody, err := c.doPost(getTaskUrl, body)
	if err != nil {
		return tr, fmt.Errorf("getTask doPost: %w", err)
	}
	err = json.Unmarshal(reqBody, &tr)
	if err != nil {
		return tr, fmt.Errorf("getTask Unmarshal req: %w", err)
	}

	return tr, nil
}

func (c *Client) ReportTask(taskReq schema.ReportTaskReq) (schema.Req, error) {
	var tr schema.Req
	body, err := json.Marshal(taskReq)
	if err != nil {
		return tr, fmt.Errorf("getTask marshal err %w", err)
	}

	reqBody, err := c.doPost(reportTaskUrl, body)
	if err != nil {
		return tr, fmt.Errorf("getTask doPost: %w", err)
	}
	err = json.Unmarshal(reqBody, &tr)
	if err != nil {
		return tr, fmt.Errorf("getTask Unmarshal req: %w", err)
	}

	return tr, nil
}

type taskWorker interface {
	DoTask(task schema.Task) (string, error)
}

func (c *Client) ListeningTasks(ctx context.Context, taskType schema.TaskType, taskWorker taskWorker, repeatTime time.Duration) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				{
					log.Println("stopping listen tasks")
					return
				}
			default:
				{
					time.Sleep(repeatTime)
					task, err := c.GetTask(schema.GetTaskReq{
						TaskType: taskType,
					})
					if err != nil {
						log.Printf("listen %d err: %s", taskType, err.Error())
						continue
					}
					if task.Data.Id == 0 {
						continue
					}
					msg, err := taskWorker.DoTask(task.Data)
					if err != nil {
						strError := fmt.Sprintf("listeningTask: %s", err.Error())
						log.Println(strError)
						_, err = c.ReportTask(schema.ReportTaskReq{
							TaskId:  task.Data.Id,
							Status:  schema.TaskStatusError,
							TextMsg: strError,
						})
						if err != nil {
							log.Printf("can't report err: %s", err.Error())
						}
					}
					_, _ = c.ReportTask(schema.ReportTaskReq{
						TaskId:  task.Data.Id,
						Status:  schema.TaskStatusDone,
						TextMsg: msg,
					})
				}
			}
		}
	}()
}

func (c *Client) doPost(url string, body []byte) ([]byte, error) {
	client := &http.Client{
		Timeout: c.timeout,
	}
	myurl := fmt.Sprintf("%s%s", c.addr, url)
	req, err := http.NewRequest("POST", myurl, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("doPost NewRequest err %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("secret", c.secret)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doPost http request %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("doPost some error status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("doPost read body %w", err)
	}
	return respBody, nil
}
