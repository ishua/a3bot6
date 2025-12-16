package taskmng

import "strings"

func (m *Mng) createFinanceTask(dialogId int64, userText string) (string, error) {
	words := strings.Split(userText, " ")
	return "не реализовано", nil
}
