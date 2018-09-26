package logruzio

import (
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	endpoint = "listener.logz.io:5050"
	proto    = "tcp"
)

// HookOpts represents Logrus Logzio hook options
type HookOpts struct {
	Conn      io.Writer
	Context   logrus.Fields
	Formatter logrus.Formatter
}

// Hook represents a Logrus Logzio hook
type Hook struct {
	hookOpts HookOpts
}

// New creates a default Logzio hook.
// What it does is taking `token` and `appName` and attaching them to the log data.
// In addition, it sets a connection to the Logzio's Logstash endpoint.
// If the connection fails, it returns an error.
//
// To set more advanced configurations, initialize the hook in the following way:
//
// hook := &Hook{HookOpts{
//		Conn: myConn,
//		Context: logrus.Fields{...},
//		Formatter: myFormatter{}
// }
func New(token string, appName string, ctx logrus.Fields) (*Hook, error) {
	return newHook(token, appName, ctx, 0)
}

// NewWithTimeout does the same as New but takes a timeout
//
// The timeout is used with net.DialTimeout so that you get more control of connection issues
func NewWithTimeout(token string, appName string, ctx logrus.Fields, timeout time.Duration) (*Hook, error) {
	return newHook(token, appName, ctx, timeout)
}

func newHook(token, appName string, ctx logrus.Fields, timeout time.Duration) (*Hook, error) {
	opts := HookOpts{Context: logrus.Fields{}}

	opts.Context["token"] = token
	opts.Context["type"] = appName
	opts.Context["meta"] = ctx
	opts.Formatter = &SimpleFormatter{}

	var conn io.Writer
	var err error
	conn, err = net.DialTimeout(proto, endpoint, timeout)
	if err != nil {
		return nil, err
	}
	opts.Conn = conn

	return &Hook{opts}, nil
}

// Fire writes `entry` to Logzio
func (h *Hook) Fire(entry *logrus.Entry) error {
	// Add in context fields.
	for k, v := range h.hookOpts.Context {
		// We don't override fields that are already set
		if _, ok := entry.Data[k]; !ok {
			entry.Data[k] = v
		}
	}

	dataBytes, err := h.hookOpts.Formatter.Format(entry)
	if err != nil {
		return err
	}

	if _, err = h.hookOpts.Conn.Write(dataBytes); err != nil {
		return err
	}

	return nil
}

// Levels returns logging levels
func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
