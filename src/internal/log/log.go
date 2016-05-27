package log

import (
	"io"

	"github.com/Sirupsen/logrus"
)

type Fields logrus.Fields

var (
	Info      = logrus.Info
	Debug     = logrus.Debug
	Error     = logrus.Error
	Panic     = logrus.Panic
	Fatal     = logrus.Fatal
	SetOutput = logrus.SetOutput
)

func (f Fields) Info(msg string)  { logrus.WithFields(logrus.Fields(f)).Info(msg) }
func (f Fields) Debug(msg string) { logrus.WithFields(logrus.Fields(f)).Debug(msg) }
func (f Fields) Error(msg string) { logrus.WithFields(logrus.Fields(f)).Error(msg) }
func (f Fields) Fatal(msg string) { logrus.WithFields(logrus.Fields(f)).Fatal(msg) }
func (f Fields) Panic(msg string) { logrus.WithFields(logrus.Fields(f)).Panic(msg) }
func EnableDebug()                { logrus.SetLevel(logrus.DebugLevel) }
func Writer() io.WriteCloser      { return logrus.StandardLogger().Writer() }
