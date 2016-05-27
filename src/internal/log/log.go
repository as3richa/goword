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
	SetOutput = logrus.SetOutput
)

func (f Fields) Info(msg string)  { logrus.WithFields(logrus.Fields(f)).Info(msg) }
func (f Fields) Debug(msg string) { logrus.WithFields(logrus.Fields(f)).Debug(msg) }
func (f Fields) Error(msg string) { logrus.WithFields(logrus.Fields(f)).Error(msg) }
func EnableDebug()                { logrus.SetLevel(logrus.DebugLevel) }
func Writer() io.WriteCloser      { return logrus.StandardLogger().Writer() }
