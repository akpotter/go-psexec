package main

import (
	"crypto/rsa"
	"github.com/ayufan/golang-kardianos-service"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/tylerb/graceful.v1"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/golang-devops/go-psexec/shared"
)

type app struct {
	logger          service.Logger
	gracefulTimeout time.Duration
	privateKey      *rsa.PrivateKey
	srv             *graceful.Server

	h *handler

	watcherPublicKeys *fsnotify.Watcher
}

func (a *app) registerInterruptSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c { //for sig := range c {
			if a.srv != nil {
				a.logger.Warning("Interrupt signal received, stopping gracefully")

				c2 := make(chan bool, 1)
				go func() {
					a.srv.Stop(a.gracefulTimeout)
					c2 <- true
				}()

				select {
				case <-c2:
					a.logger.Warning("Graceful shutdown complete")
					os.Exit(0)
				case <-time.After(a.gracefulTimeout):
					a.logger.Warning("Normal timeout, forcefully exiting")
					os.Exit(1)
				}
			}
		}
	}()
}

func (a *app) Run(logger service.Logger) {
	a.logger = logger
	a.gracefulTimeout = 30 * time.Second //Because a command could be busy executing

	pvtKey, err := shared.ReadPemKey(*serverPemFlag)
	if err != nil {
		logger.Errorf("Cannot read server pem file, error: %s. Exiting server.", err.Error())
		return
	}
	a.privateKey = pvtKey

	a.h = &handler{logger, a.privateKey, nil}

	allowedKeys, err := shared.LoadAllowedPublicKeysFile(*allowedPublicKeysFileFlag)
	if err != nil {
		logger.Errorf("Cannot read allowed public keys, error: %s. Exiting server.", err.Error())
		return
	}
	if len(allowedKeys) == 0 {
		logger.Errorf("Allowed public key file '%s' was read but contains no keys. Exiting server.", *allowedPublicKeysFileFlag)
		return
	}

	a.setAllowedPublicKeys(allowedKeys)

	watcher, err := shared.StartWatcher(*allowedPublicKeysFileFlag, a)
	if err != nil {

	} else {
		a.watcherPublicKeys = watcher
	}

	e := echo.New()
	e.Debug()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	t := &htmlTemplateRenderer{}
	e.SetRenderer(t)

	// Unrestricted group
	e.Post("/token", a.h.handleGenerateTokenFunc)
	e.Get("/webui", a.h.handleWebUIFunc)

	// Restricted group
	r := e.Group("/auth")
	r.Use(GetClientPubkey())
	r.Post("/exec", a.h.handleExecFunc)

	a.logger.Infof("Now serving on '%s'", *addressFlag)

	a.srv = &graceful.Server{
		Timeout: a.gracefulTimeout,
		Server:  e.Server(*addressFlag),
	}

	a.registerInterruptSignal()

	a.srv.ListenAndServe()
	if err != nil {
		if !strings.Contains(err.Error(), "closed network connection") {
			logger.Errorf("Unable to ListenAndServe, error: %s", err.Error())
		}
	}
}

func (a *app) OnStop() {
	if a.watcherPublicKeys != nil {
		a.watcherPublicKeys.Close()
	}

	a.srv.Stop(a.gracefulTimeout)
	<-a.srv.StopChan()
}
