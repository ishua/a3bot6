package functions

import "fmt"

type Mng struct {
	repo repo
}

type repo interface {
	DeleteAllTasks() error
	DeleteAllDialogs() error
}

func NewMng(repo repo) *Mng {
	return &Mng{repo: repo}
}

func (mng *Mng) DeleteAll() error {
	err := mng.repo.DeleteAllTasks()
	if err != nil {
		return fmt.Errorf("deleteAll tasks Error: %s", err.Error())
	}

	err = mng.repo.DeleteAllDialogs()
	if err != nil {
		return fmt.Errorf("deleteAll dialogs Error: %s", err.Error())
	}
	return nil
}
