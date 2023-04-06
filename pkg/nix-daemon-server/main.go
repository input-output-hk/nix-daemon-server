package main

import (
	"github.com/alexflint/go-arg"
	"github.com/gliderlabs/ssh"
	"go.uber.org/zap"
)

func main() {
	c := newConfig()
	arg.MustParse(c)

	p := proxy{config: c, sessions: make(chan bool, c.MaxSessions)}
	p.setupLog()
	p.allowedKeys = p.syncAllowedKeys()

	p.log.Info("Starting server",
		zap.String("address", c.ListenAddr),
		zap.String("host key", c.HostKeyPath),
		zap.String("log level", c.LogLevel),
		zap.String("host mode", c.LogMode),
		zap.Int64("max session", c.MaxSessions),
		zap.Duration("new connection timeout", c.NewConnTime),
		zap.Duration("github sync timer", c.GHSyncTime),
		zap.String("github team", c.GHTeam),
		zap.String("github organization", c.GHOrg),
		zap.String("github token path", c.GHTokenPath),
	)

	// TODO: add connection timeouts
	if err := ssh.ListenAndServe(c.ListenAddr, p.handler, ssh.HostKeyFile(c.HostKeyPath), ssh.PublicKeyAuth(p.auth)); err != nil {
		p.log.Fatal("Failed to start server", zap.Error(err))
	}
}
