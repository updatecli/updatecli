package log

import (
	"github.com/sirupsen/logrus"
)

// Dependencies such as helm or sops log through the global logrus logger,
// routing it through our output keeps their messages in captures and
// formatted like ours.
func init() {
	logrus.SetOutput(fanout{})
	logrus.SetFormatter(logrusFormatter{})
	setLogrusLevel(InfoLevel)
}

type logrusFormatter struct{}

func (logrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return render(fromLogrusLevel(entry.Level), entry.Message), nil
}

func fromLogrusLevel(l logrus.Level) Level {
	switch {
	case l <= logrus.ErrorLevel:
		return ErrorLevel
	case l == logrus.WarnLevel:
		return WarnLevel
	case l == logrus.InfoLevel:
		return InfoLevel
	default:
		return DebugLevel
	}
}

func setLogrusLevel(l Level) {
	switch {
	case l >= ErrorLevel:
		logrus.SetLevel(logrus.ErrorLevel)
	case l >= WarnLevel:
		logrus.SetLevel(logrus.WarnLevel)
	case l >= InfoLevel:
		logrus.SetLevel(logrus.InfoLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}
