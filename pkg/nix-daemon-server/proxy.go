package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	xssh "golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

type proxy struct {
	config      *config
	log         *zap.Logger
	sessions    chan bool
	allowedKeys *sync.Map
}

func (p *proxy) setupLog() {
	lvl := zap.NewAtomicLevel()
	if err := lvl.UnmarshalText([]byte(p.config.LogLevel)); err != nil {
		panic(err)
	}
	development := p.config.LogMode == "development"
	encoding := "json"
	encoderConfig := zap.NewProductionEncoderConfig()
	if development {
		encoding = "console"
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	l := zap.Config{
		Level:             lvl,
		Development:       development,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          &zap.SamplingConfig{Initial: 1, Thereafter: 2},
		Encoding:          encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	var err error
	p.log, err = l.Build()
	if err != nil {
		panic(err)
	}
}

func (p *proxy) handler(s ssh.Session) {
	p.log.Debug("new session")
	defer func() { p.log.Debug("end session") }()
	select {
	case p.sessions <- true:
		p.log.Debug("got session")
		defer func() { <-p.sessions; p.log.Debug("return session") }()
	case <-time.After(p.config.NewConnTime):
		_, _ = io.WriteString(s.Stderr(), "Too many connections\n")
		_ = s.Exit(1)
	}

	cmd := exec.Command("go", "run", "./pkg/nix-daemon-protocol", "--stdio")
	cmd.Env = append(os.Environ(),
		"GITHUB_USER=%s"+s.Context().Value("GITHUB_USER").(string),
		"SSH_USER=%s"+s.Context().User(),
		"PUB_KEY_HASH=%s"+xssh.FingerprintSHA256(s.PublicKey()),
	)
	cmd.Stderr = s.Stderr()
	cmd.Stdin = s
	cmd.Stdout = s
	if err := cmd.Run(); err != nil {
		p.log.Error("nix-daemon failed", zap.Error(err), zap.String("state", cmd.ProcessState.String()))
		_ = s.Exit(1)
	}
	p.log.Debug("nix-daemon returned", zap.String("state", cmd.ProcessState.String()))
	_ = s.Exit(0)
}

func (p *proxy) auth(ctx ssh.Context, key ssh.PublicKey) bool {
	allow := false
	p.allowedKeys.Range(func(_, mk any) bool {
		if mwk, ok := mk.(memberWithKey); !ok {
			p.log.DPanic("memberWithKey assertion failed")
		} else if ssh.KeysEqual(key, mwk.key) {
			p.log.Info("login allowed", zap.String("login", mwk.login), zap.String("key", xssh.FingerprintSHA256(key)))
			ctx.SetValue("GITHUB_USER", mwk.login)
			allow = true
			return false
		}

		return true
	})

	if !allow {
		p.log.Info("login denied", zap.String("key", xssh.FingerprintSHA256(key)))
	}

	return allow
}

func (p *proxy) syncAllowedKeys() *sync.Map {
	m := &sync.Map{}
	if err := p.syncGithub(m); err != nil {
		p.log.Fatal("initially syncing github", zap.Error(err))
	}

	go func() {
		timer := time.Tick(p.config.GHSyncTime)
		for range timer {
			if err := p.syncGithub(m); err != nil {
				p.log.Fatal("while syncing github", zap.Error(err))
			}
		}
	}()

	return m
}

// TODO: handle more than 100 members
type KeyQuery struct {
	Organization struct {
		Login githubv4.String
		Team  struct {
			Name    githubv4.String
			Members struct {
				Nodes []struct {
					Login      githubv4.String
					PublicKeys struct {
						Nodes []struct {
							Key githubv4.String
						}
					} `graphql:"publicKeys(first:100)"`
				}
			} `graphql:"members(first:100)"`
		} `graphql:"team(slug:$team)"`
	} `graphql:"organization(login:$org)"`
}

func (p *proxy) syncGithub(keys *sync.Map) error {
	p.log.Debug("fetching github keys")
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: p.config.Token(p.log)},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	query := KeyQuery{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Query(ctx, &query, map[string]interface{}{
		"org":  githubv4.String(p.config.GHOrg),
		"team": githubv4.String(p.config.GHTeam),
	}); err != nil {
		return errors.WithMessage(err, "while querying github")
	}

	seen := map[string]bool{}

	for _, member := range query.Organization.Team.Members.Nodes {
		login := string(member.Login)
		for _, key := range member.PublicKeys.Nodes {
			publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key.Key))
			if err != nil {
				p.log.Error("while parsing key", zap.Error(err))
			}
			keyHash := xssh.FingerprintSHA256(publicKey)
			seen[keyHash] = true
			_, loaded := keys.LoadOrStore(keyHash, memberWithKey{login, publicKey})
			if !loaded {
				p.log.Debug("stored new key", zap.String("key", keyHash), zap.String("login", login))
			}
		}
	}

	keys.Range(func(key, value any) bool {
		if _, ok := seen[key.(string)]; !ok {
			keys.Delete(key)
			p.log.Debug("deleted old key", zap.String("key", key.(string)), zap.String("login", value.(memberWithKey).login))
		}
		return true
	})

	return nil
}

type memberWithKey struct {
	login string
	key   ssh.PublicKey
}
