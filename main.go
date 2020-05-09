package main

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	log = logrus.New()
)

func init() {
	// init logger
	log.SetFormatter(&prefixed.TextFormatter{
		ForceColors:      false,
		DisableColors:    false,
		ForceFormatting:  true,
		DisableTimestamp: true,
		DisableUppercase: false,
		FullTimestamp:    false,
		TimestampFormat:  "",
		DisableSorting:   false,
		QuoteEmptyFields: false,
		QuoteCharacter:   "",
		SpacePadding:     0,
		Once:             sync.Once{},
	})

	log.SetOutput(os.Stdout)
	log.Level = logrus.TraceLevel
}

func getImportedFile() (string, error) {
	// determine file that was imported
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) < 2 {
			continue
		}

		switch strings.ToLower(pair[0]) {
		case "sonarr_episodefile_path", "radarr_moviefile_path", "lidarr_trackfile_path":
			return pair[1], nil
		default:
			break
		}
	}

	return "", errors.New("could not find imported env var")
}

func getEventType() (string, error) {
	// determine event type
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) < 2 {
			continue
		}

		switch strings.ToLower(pair[0]) {
		case "sonarr_eventtype", "radarr_eventtype", "lidarr_eventtype":
			return pair[1], nil
		default:
			break
		}
	}

	return "", errors.New("could not find eventtype env var")
}

func resetFileTime(fp string) error {
	// set logger
	log := log.WithField("file", fp)

	// stat file
	file, err := os.Stat(fp)
	if err != nil {
		return errors.WithMessage(err, "could not stat file")
	}

	// get current times
	mt := file.ModTime()
	nt := time.Now().Local()

	// change mod-time
	if err := os.Chtimes(fp, nt, nt); err != nil {
		return errors.WithMessage(err, "could not chtime file")
	}

	log.WithFields(logrus.Fields{
		"old_mod_time": mt,
		"new_mod_time": nt,
	}).Info("Reset mod-time")
	return nil
}

func main() {
	// determine event type
	et, err := getEventType()
	if err != nil {
		log.WithError(err).Fatal("Failed determining event type")
	}

	// proceed no further for tests
	if strings.EqualFold(et, "Test") {
		os.Exit(0)
	}

	// determine file that was imported
	f, err := getImportedFile()
	if err != nil {
		log.WithError(err).Fatal("Failed determining imported file")
	}

	log := log.WithField("file", f)
	log.Info("Found imported file")

	// reset mod-time for this file
	if err := resetFileTime(f); err != nil {
		log.WithError(err).Fatal("Failed resetting mod-time of imported file")
	}
}
