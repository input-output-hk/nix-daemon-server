package main

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type config struct {
	HostKeyPath string        `arg:"--host-key,env:HOST_KEY" help:"SSH server host key"`
	ListenAddr  string        `arg:"--listen,env:LISTEN_ADDR" help:"Listen on this address:port"`
	LogLevel    string        `arg:"--log-level,env:LOG_LEVEL" help:"one of debug, info, warn, error, dpanic, panic, fatal"`
	LogMode     string        `arg:"--log-mode,env:LOG_MODE" help:"development or production"`
	MaxSessions int64         `arg:"--max-sessions,env:MAX_SESSIONS" help:"maximum amount of concurrent sessions"`
	NewConnTime time.Duration `arg:"--new-connection-timeout,env:CONNECTION_TIMEOUT" help:"how long new connections may be delayed before a session becomes available"`
	GHSyncTime  time.Duration `arg:"--github-refresh-interval,env:GITHUB_REFRESH_INTERVAL" help:"synchronize allowed keys from Github every interval"`
	GHOrg       string        `arg:"--github-organization,required,env:GITHUB_ORGANIZATION" help:"organization the team is in"`
	GHTeam      string        `arg:"--github-team,required,env:GITHUB_TEAM" help:"fetch keys of the members of this team"`
	GHToken     string        `arg:"--github-token,env:GITHUB_TOKEN" help:"github token; takes precedence over the token path"`
	GHTokenPath string        `arg:"--github-token-path,env:GITHUB_TOKEN_PATH" help:"read github token from a file instead"`
}

func newConfig() *config {
	return &config{
		HostKeyPath: "./ssh_host_ed25519_key",
		ListenAddr:  "0.0.0.0:2222",
		LogLevel:    "debug",
		LogMode:     "development",
		MaxSessions: 2,
		NewConnTime: 1 * time.Second,
		GHSyncTime:  1 * time.Minute,
	}
}

var (
	buildVersion = "dev"
	buildCommit  = "dirty"
)

func (config) Version() string {
	return buildVersion + " (" + buildCommit + ")"
}

// TODO: depending on the token rotation strategy, this may be subject to race
// conditions when the token file is not replaced atomically.
// We could work around it with caching, but that will cause a lot more
// complexity so we just assume responsible usage and crash when the file is
// unreadable or incomplete.
func (c config) Token(log *zap.Logger) string {
	if c.GHToken != "" {
		return strings.TrimSpace(c.GHToken)
	}

	if c.GHTokenPath != "" {
		token, err := os.ReadFile(c.GHTokenPath)
		if err != nil {
			log.Fatal("couldn't read token", zap.Error(err), zap.String("path", c.GHTokenPath))
		}
		return strings.TrimSpace(string(token))
	} else {
		println("error: --github-token or --github-token-path is required (alternatively environment variables GITHUB_TOKEN or GITHUB_TOKEN_PATH)")
		os.Exit(1)
	}

	panic("unreachable")
}
