package controller

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Instabug/netbird-gitops/pkg/client"
	"github.com/Instabug/netbird-gitops/pkg/data"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/nikoksr/notify"
	"gopkg.in/yaml.v3"
)

const localRepoPath = "/tmp/netbird-gitops"

// Controller main logic controller
type Controller struct {
	netbirdClient client.Client
	*Options
}

// Options controller settings
type Options struct {
	GitRepoURL      string
	GitRelativePath string
	GitBranch       string
	GitAuth         transport.AuthMethod
	NetBirdToken    string
	NetBirdAPI      string
	SyncOnceAndExit bool
	PollFrequency   time.Duration
}

// NewController init
func NewController(opts Options) *Controller {
	return &Controller{
		Options: &opts,
	}
}

// Start blocks and begins logic
func (c *Controller) Start(ctx context.Context) error {
	// Init NetBird Client
	c.netbirdClient = *client.NewClient(c.NetBirdAPI, c.NetBirdToken, true)
	// Clone initial repository and handle sync logic
	slog.Info("Cloning repository")
	os.RemoveAll(localRepoPath)
	repo, err := git.PlainCloneContext(ctx, localRepoPath, false, &git.CloneOptions{
		URL:           c.GitRepoURL,
		ReferenceName: plumbing.NewBranchReferenceName(c.GitBranch),
		RemoteName:    "origin",
		Auth:          c.GitAuth,
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("Failed to initialize repo pull: %w", err)
	}

	defer func() {
		os.RemoveAll(localRepoPath)
	}()

	cfg, err := c.getCombinedConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)

	}

	if err := c.doSync(ctx, cfg, !c.SyncOnceAndExit && cfg.Config.AutoSync == "manual"); err != nil {
		slog.Error("Failed to sync", "err", err)
		notify.Send(ctx, "Sync failed", fmt.Sprintf("Failed to do initial sync due to error: %s", err.Error()))
	}

	if c.SyncOnceAndExit {
		return nil
	}
	// Start polling loop for changes
	latestHead, err := repo.Head()
	if err != nil {
		return fmt.Errorf("Failed to get repo HEAD: %w", err)
	}
	latestCommit, err := repo.CommitObject(latestHead.Hash())
	if err != nil {
		return fmt.Errorf("Failed to get repo HEAD: %w", err)
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("Failed to get repo WorkTree: %w", err)
	}

	slog.Info("Starting poll loop")
	t := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
		slog.Info("Pulling changes")
		err = workTree.PullContext(ctx, &git.PullOptions{
			RemoteName:    "origin",
			RemoteURL:     c.GitRepoURL,
			ReferenceName: plumbing.NewBranchReferenceName(c.GitBranch),
			Auth:          c.GitAuth,
			SingleBranch:  true,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			slog.Error("Failed to pull repo", "err", err)
			notify.Send(ctx, "Pull failed", fmt.Sprintf("Failed to pull repository %s with error: %s", c.GitRepoURL, err.Error()))
			continue
		}

		curHead, err := repo.Head()
		if err != nil {
			slog.Error("Failed to get repo HEAD", "err", err)
			notify.Send(ctx, "Get repo HEAD failed", fmt.Sprintf("Failed to get repository %s HEAD with error: %s", c.GitRepoURL, err.Error()))
			continue
		}
		curCommit, err := repo.CommitObject(latestHead.Hash())
		if err != nil {
			slog.Error("Error getting latest commit")
			notify.Send(ctx, "Get repo HEAD failed", fmt.Sprintf("Failed to get repository %s HEAD with error: %s", c.GitRepoURL, err.Error()))
			continue
		}

		cfg, err := c.getCombinedConfig()
		if err != nil {
			slog.Error("error loading config", "err", err)
			notify.Send(ctx, "Failed to load config", fmt.Sprintf("Failed to load config from %s/%s with error: %s", c.GitRepoURL, c.GitRelativePath, err.Error()))
			continue
		}

		dryRun := true
		switch cfg.Config.AutoSync {
		case "enforce":
			dryRun = false
		case "update":
			iter, err := repo.Log(&git.LogOptions{
				Order: git.LogOrderCommitterTime,
				PathFilter: func(s string) bool {
					return strings.HasPrefix(s, strings.TrimPrefix(c.GitRelativePath, "/"))
				},
			})
			cmt, err := iter.Next()
			slog.Debug("Latest hash", "hash", latestHead.Hash().String())
			for err == nil && cmt.Hash.String() != latestHead.Hash().String() {
				ancestor, err := latestCommit.IsAncestor(cmt)
				if err != nil {
					slog.Error("Error checking commit ancestry", "ancestor", latestCommit.Hash.String(), "descendant", cmt.Hash.String(), "err", err)
				}
				if !ancestor {
					break
				}
				slog.Debug("Checking commit", "hash", cmt.Hash.String())
				fs, err := cmt.Stats()
				if err != nil {
					slog.Error("Error checking commit stats", "commit", cmt.Hash.String(), "err", err)
				}
				if len(fs) > 0 {
					dryRun = false
				}
				cmt, err = iter.Next()
			}
			if err != nil {
				slog.Error("Error iterating commits", "err", err)
				continue
			}
		case "manual":
			dryRun = true
		}

		if err := c.doSync(ctx, cfg, dryRun); err != nil {
			notify.Send(ctx, "Sync failed", fmt.Sprintf("Failed to sync %s/%s with error: %s", c.GitRepoURL, c.GitRelativePath, err.Error()))
			slog.Error("Failed to sync", "err", err)
		}

		latestHead = curHead
		latestCommit = curCommit
	}
}

func (c *Controller) getCombinedConfig() (*data.CombinedConfig, error) {
	localPath := path.Join(localRepoPath, c.GitRelativePath)
	files, err := os.ReadDir(localPath)
	if err != nil {
		return nil, err
	}

	var filesBytes [][]byte
	for _, f := range files {
		ext := path.Ext(f.Name())
		if ext != ".yaml" && ext != ".yml" {
			slog.Info("ignoring file", "name", f.Name())
			continue
		}
		slog.Info("found file", "name", f.Name())

		fileBytes, err := os.ReadFile(path.Join(localPath, f.Name()))
		if err != nil {
			return nil, err
		}

		filesBytes = append(filesBytes, fileBytes)
	}

	totalBytes := bytes.Join(filesBytes, []byte("\n"))

	cfg := &data.CombinedConfig{}
	err = yaml.Unmarshal(totalBytes, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
