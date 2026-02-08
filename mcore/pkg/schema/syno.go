package schema

type SynoTaskCmd string

const (
	SynoTaskCmdAdd    SynoTaskCmd = "add"
	SynoTaskCmdList   SynoTaskCmd = "list"
	SynoTaskCmdDelete SynoTaskCmd = "delete"
)

type SynoCategory string

const (
	SynoCategoryMovie         SynoCategory = "movie"
	SynoCategoryCartoon       SynoCategory = "cartoon"
	SynoCategoryShows         SynoCategory = "shows"
	SynoCategoryAudiobook     SynoCategory = "audiobook"
	SynoCategoryOther         SynoCategory = "other"
	SynoCategoryShowsCartoons SynoCategory = "shows_cartoons"
)

type TaskSyno struct {
	Command    SynoTaskCmd  `json:"command"`
	Category   SynoCategory `json:"category"`
	TorrentUrl string       `json:"torrentUrl"`
	TaskId     string       `json:"taskId"`
}
