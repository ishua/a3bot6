package dialogmng

import (
	"github.com/ishua/a3bot6/mcore/pkg/schema"
)

type DialogMng struct {
	repo repo
}

func NewDialogMng(repo repo) *DialogMng {
	return &DialogMng{repo: repo}
}

type repo interface {
	AddDialog(dialog schema.Dialog) (int64, error)
}

func (d *DialogMng) Create(m schema.Message) (int64, error) {
	return d.repo.AddDialog(schema.Dialog{
		Key:          schema.GenerateKey(m),
		DialogStatus: schema.DialogStatusBegin,
		Messages:     []schema.Message{m},
	})
}
