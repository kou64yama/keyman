package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"keyman"
	"keyman/pb"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/dgraph-io/badger"
	"google.golang.org/grpc"
)

type badgerLogger struct {
	logger *log.Logger
}

func (l *badgerLogger) Errorf(fmt string, args ...interface{}) {
	l.logger.Printf("ERROR "+fmt, args...)
}

func (l *badgerLogger) Warningf(fmt string, args ...interface{}) {
	l.logger.Printf("WARNING "+fmt, args...)
}

func (l *badgerLogger) Infof(fmt string, args ...interface{}) {
	l.logger.Printf("INFO "+fmt, args...)
}

func (l *badgerLogger) Debugf(fmt string, args ...interface{}) {
	l.logger.Printf("DEBUG "+fmt, args...)
}

func daemon(c *config) (*exec.Cmd, error) {
	cmd := exec.Command(os.Args[0], "-p", c.dir, "-s", c.socket)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	s := bufio.NewScanner(stdout)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "Listen on ") {
			return cmd, nil
		}
	}

	return cmd, cmd.Wait()
}

func server(c *config) error {
	fmt.Printf("Keyman version %s\n", version)
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.socket), 0755); err != nil {
		return err
	}
	ln, err := net.Listen("unix", c.socket)
	if err != nil {
		return err
	}
	defer ln.Close()
	logger := &keyman.Logger{
		ErrorLog:   log.New(os.Stderr, "", 0),
		WarningLog: log.New(os.Stdout, "", 0),
		InfoLog:    log.New(ioutil.Discard, "", 0),
		DebugLog:   log.New(ioutil.Discard, "", 0),
	}
	if c.debug {
		logger.DebugLog = log.New(os.Stdout, "", 0)
	}

	opt := badger.DefaultOptions(c.dir).WithLogger(logger)
	db, err := badger.Open(opt)
	if err != nil {
		return err
	}
	defer db.Close()
	fmt.Fprintf(os.Stdout, "Open %s\n", c.dir)

	srv := grpc.NewServer()
	pb.RegisterKeymanServer(srv, keyman.NewServer(logger, db))

	closeIdleConnections := make(chan interface{}, 1)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)

		received := <-sig
		fmt.Fprintf(os.Stderr, "Shutdown: %s\n", received)
		srv.GracefulStop()
		close(closeIdleConnections)
	}()

	fmt.Fprintf(os.Stdout, "Listen on %s\n", ln.Addr())
	if err := srv.Serve(ln); err != nil {
		return err
	}
	<-closeIdleConnections

	db.Close()
	ln.Close()
	return nil
}
