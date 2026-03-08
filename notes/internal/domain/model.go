package domain

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ishua/a3bot6/mcore/pkg/logger"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"github.com/ishua/a3bot6/notes/internal/clients/gitapi"
)

const (
	inboxPath = "main.markdown"
)

type Model struct {
	gitClient gitClient
}

type gitClient interface {
	Pull() error
	GetRepoPath() string
	CommitAndPush(path []string) error
}

func NewModel(gitClient *gitapi.GitClient) *Model {
	return &Model{
		gitClient,
	}
}

func (m *Model) DoTask(task schema.Task) schema.ReportTaskReq {
	logger.Debug("starting task")
	err := m.gitClient.Pull()
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("execute pull repo: %v", err),
		}
	}
	if len(task.TaskData.Health) > 0 {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusDone,
			TextMsg: "note is healthy",
		}
	}
	switch task.TaskData.Tn.Command {
	case schema.TaskNoteCmdPull:
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusDone,
			TextMsg: "Successfully pulled",
		}
	case schema.TaskNoteReadInbox:
		return m.readInbox(task.Id)
	case schema.TaskNoteCmdAddDiary:
		return m.addAddDiary(task)
	case schema.TaskNoteCmdAddInbox:
		return m.addAddInbox(task)
	}
	return schema.ReportTaskReq{
		TaskId:  task.Id,
		Status:  schema.TaskStatusError,
		TextMsg: fmt.Sprintf("notes.model unknown command: %s", task.TaskData.Tn.Command),
	}
}

func (m *Model) addAddDiary(task schema.Task) schema.ReportTaskReq {
	addText := task.TaskData.Tn.AddText
	if len(addText) == 0 {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: "add text is empty",
		}
	}
	diaryRows, err := m.readFile(getDiaryPath())
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("read diary file: %v", err),
		}
	}
	newStrings := []string{}

	h2 := "## " + time.Now().Format("0201")
	if len(diaryRows) == 0 || isH2NotExist(h2, diaryRows) {
		newStrings = append(newStrings, h2)
	}

	newStrings = append(newStrings, "- "+addText)
	err = m.addRowToFile(getDiaryPath(), newStrings)
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("add diary to file: %v", err),
		}
	}

	err = m.gitClient.CommitAndPush([]string{getDiaryPath()})
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("commit and push diary: %v", err),
		}
	}

	return schema.ReportTaskReq{
		TaskId:  task.Id,
		Status:  schema.TaskStatusDone,
		TextMsg: "text add to diary",
	}
}

func isH2NotExist(h2 string, diaryRows []string) bool {
	for i := len(diaryRows) - 1; i >= 0; i-- {
		if diaryRows[i] == h2 {
			return false
		}
	}
	return true
}

func (m *Model) addAddInbox(task schema.Task) schema.ReportTaskReq {
	addText := task.TaskData.Tn.AddText
	if len(addText) == 0 {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: "add text is empty",
		}
	}
	line := fmt.Sprintf("  - %s", addText)
	err := m.addRowToFile(inboxPath, []string{line})
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("add to inbox: %v", err),
		}
	}

	err = m.gitClient.CommitAndPush([]string{inboxPath})
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  task.Id,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("commit and push inbox: %v", err),
		}
	}

	return schema.ReportTaskReq{
		TaskId:  task.Id,
		Status:  schema.TaskStatusDone,
		TextMsg: "text add to inbox",
	}
}

func (m *Model) readInbox(taskId int64) schema.ReportTaskReq {
	rows, err := m.readFile(inboxPath)
	if err != nil {
		return schema.ReportTaskReq{
			TaskId:  taskId,
			Status:  schema.TaskStatusError,
			TextMsg: fmt.Sprintf("read inbox: %v", err),
		}
	}

	var ret string

	for idx, line := range rows {
		if len(rows) == idx+1 {
			ret += line
			break
		}
		ret += fmt.Sprintf("%s\n", line)
	}

	return schema.ReportTaskReq{
		TaskId:  taskId,
		Status:  schema.TaskStatusDone,
		TextMsg: ret,
	}
}

func (m *Model) readFile(filePath string) ([]string, error) {
	path := filepath.Join(m.gitClient.GetRepoPath(), filePath)

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("model open inbox: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var ret []string
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("model scaner: %w", err)
	}

	return ret, nil
}

func (m *Model) addRowToFile(filePath string, rows []string) error {
	path := filepath.Join(m.gitClient.GetRepoPath(), filePath)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("addRowToFile OpenFile: %w", err)
	}
	for _, r := range rows {
		_, err = f.WriteString(r + "\n")
		if err != nil {
			return fmt.Errorf("addRowToFile WriteString: %w", err)
		}
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("addRowToFile file.Close: %w", err)
	}
	return nil
}

func getDiaryPath() string {
	now := time.Now()
	quarter := (int(now.Month())-1)/3 + 1 // 1,2,3,4
	return fmt.Sprintf("Diary/5BX %d%02d.markdown", now.Year(), quarter)
}
