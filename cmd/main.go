package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Instabug/netbird-gitops/pkg/controller"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func envDefault(envVar, def string) string {
	if v, ok := os.LookupEnv(envVar); ok {
		return v
	}
	return def
}

var (
	gitRepoURL            = flag.String("git-repo-url", os.Getenv("GIT_REPO_URL"), "Git Repo URL (ssh/https) (Required)")
	gitRelativePath       = flag.String("git-relative-path", os.Getenv("GIT_RELATIVE_PATH"), "Relative path of NetBird configuration within the git repo")
	gitBranch             = flag.String("git-branch", envDefault("GIT_BRANCH", "main"), "Name of branch to pull changes from")
	gitAuthMethod         = flag.String("git-auth-method", envDefault("GIT_AUTH_METHOD", "none"), "basic (username-password/access token), or ssh (private key), or none")
	gitUsername           = flag.String("git-username", os.Getenv("GIT_USERNAME"), "git basic auth username, must be defined if --git-auth-method is basic")
	gitPassword           = flag.String("git-password", os.Getenv("GIT_PASSWORD"), "git basic auth password, must be defined if --git-auth-method is basic")
	gitPrivateKeyPath     = flag.String("git-private-key-path", os.Getenv("GIT_PRIVATE_KEY_PATH"), "git SSH private key path, must be defined if --git-auth-method is ssh")
	gitPrivateKeyPassword = flag.String("git-private-key-password", os.Getenv("GIT_PRIVATE_KEY_PASSWORD"), "git SSH private key password (if any)")
	syncFrequency         = flag.Duration("sync-frequency", time.Minute, "Time between syncs")
	netbirdToken          = flag.String("netbird-token", os.Getenv("NETBIRD_TOKEN"), "NetBird Management API token")
	netbirdManagementAPI  = flag.String("netbird-mgmt-api", os.Getenv("NETBIRD_MANAGEMENT_API"), "NetBird Management API URL")
	logLevel              = flag.String("log-level", os.Getenv("LOG_LEVEL"), "Log level (debug, info, warn, error)")
	syncExit              = flag.Bool("sync-and-exit", false, "Force sync once and exit")
	notifyServicesPath    = flag.String("notify-services-path", "notify.yaml", "Path to notification services configuration yaml")
)

func main() {
	flag.Parse()

	level := slog.LevelInfo
	if *logLevel != "" {
		if err := level.UnmarshalText([]byte(*logLevel)); err != nil {
			slog.Warn("Error setting log level", "err", err)
		}
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})))

	if *netbirdToken == "" {
		flag.PrintDefaults()
		fmt.Println("--netbird-token is required")
		os.Exit(1)
	}

	if *netbirdManagementAPI == "" {
		flag.PrintDefaults()
		fmt.Println("--netbird-mgmt-api is required")
		os.Exit(1)
	}

	if *gitRepoURL == "" {
		flag.PrintDefaults()
		fmt.Println("--git-repo-url is required")
		os.Exit(1)
	}

	var gitAuth transport.AuthMethod
	switch *gitAuthMethod {
	case "basic":
		if *gitUsername == "" || *gitPassword == "" {
			flag.PrintDefaults()
			fmt.Println("--git-auth-method is basic, but one of --git-username or --git-password is empty")
			os.Exit(1)
		}
		gitAuth = &http.BasicAuth{
			Username: *gitUsername,
			Password: *gitPassword,
		}
		break
	case "ssh":
		if *gitPrivateKeyPath == "" {
			flag.PrintDefaults()
			panic("--git-auth-method is ssh, but --git-private-key-path is empty")
		}
		_, err := os.Stat(*gitPrivateKeyPath)
		if err != nil {
			fmt.Println("private key not found")
			os.Exit(1)
		}
		publicKeys, err := ssh.NewPublicKeysFromFile("git", *gitPrivateKeyPath, *gitPrivateKeyPath)
		if err != nil {
			slog.Error("Error loading private key", "path", gitPrivateKeyPath, "err", err, "using_password", *gitPrivateKeyPassword != "")
			fmt.Println("Could not load private key")
			os.Exit(1)
		}
		gitAuth = publicKeys
		break
	case "none":
		break
	default:
		flag.PrintDefaults()
		fmt.Println("Unknown --git-auth-method")
		os.Exit(1)
	}

	err := setupNotifiers()
	if err != nil {
		slog.Warn("Error setting up notifications", "err", err)
	}

	ctrl := controller.NewController(controller.Options{
		GitRepoURL:      *gitRepoURL,
		GitRelativePath: *gitRelativePath,
		GitBranch:       *gitBranch,
		GitAuth:         gitAuth,
		NetBirdToken:    *netbirdToken,
		NetBirdAPI:      *netbirdManagementAPI,
		SyncOnceAndExit: *syncExit,
	})

	ctx, cancel := context.WithCancel(context.Background())

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigc
		cancel()
		slog.Warn("Signal received, shutting down")
		os.Exit(0)
	}()

	if err := ctrl.Start(ctx); err != nil {
		panic(err)
	}
	sigc <- os.Interrupt
}
