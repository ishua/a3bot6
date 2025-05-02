package gitapi

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"os"
	"time"
)

type GitClient struct {
	*git.Repository
	Token    string
	Username string
	Email    string
	RepoPath string
}

func NewClient(path string, url string, token string, username string, email string) (*GitClient, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		r, err = git.PlainClone(path, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: username, // yes, this can be anything except an empty string
				Password: token,
			},
			URL:      url,
			Progress: os.Stdout,
		})
		if err != nil {
			return nil, err
		}
	}

	return &GitClient{r, token, username, email, path}, nil
}

func (r *GitClient) Pull() error {
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("pull create worktree: %w", err)
	}
	err = w.Pull(&git.PullOptions{
		Auth: &http.BasicAuth{
			Username: r.Username, // yes, this can be anything except an empty string
			Password: r.Token,
		},
		RemoteName: "origin"})
	if err != nil {
		if err.Error() == "already up-to-date" {
			err = nil
		}
	}
	return err
}

func (r *GitClient) CommitAndPush(path []string) error {
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("commitAndPush create worktree: %w", err)
	}
	for _, p := range path {
		_, err = w.Add(p)
		if err != nil {
			return fmt.Errorf("commitAndPush file %s %w", p, err)
		}
	}

	_, err = w.Commit("fsnotes bots commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  r.Username,
			Email: r.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("commitAndPush commit %w", err)
	}

	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: r.Username,
			Password: r.Token,
		},
	})

	if err != nil {
		return fmt.Errorf("commitAndPush push %w", err)
	}

	return nil
}

func (r *GitClient) GetRepoPath() string {
	return r.RepoPath
}
