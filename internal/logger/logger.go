package logger

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type Formatter struct {
	Location  *time.Location
	Formatter *logrus.JSONFormatter
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.In(f.Location)

	return f.Formatter.Format(entry)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		logrus.Error("logger.go Error loading .env file: " + err.Error())
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		logrus.Fatalln(err)
	}

	log := logrus.New()

	log.SetFormatter(&Formatter{
		Location: location,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.DateTime,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
			},
		},
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	logger = log.WithFields(logrus.Fields{
		"project": "manga_store",
	})
}

func Info(msg string) {
	logger.Infoln(msg)
}

func Error(msg string) {
	logger.Errorln(msg)
}

func Debug(msg string) {
	logger.Debugln(msg)
}

func Trace(msg string) {
	logger.Traceln(msg)
}

func Warn(msg string) {
	logger.Warnln(msg)
}