package utils

import (
	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

// LogInit for logging
func LogInit() (*etilogger.Logger, func(), error) {
	appName := ApplicationNameKey

	// Initialize Logger
	logConfig := etilogger.DefaultProdConfig
	logConfig.DisableStacktrace = true
	logger, flush, err := etilogger.New(appName, logConfig)
	if err != nil {
		flush()
		return nil, nil, err
	}

	// Set Logging Prefix for Data Service
	logger.SetTrackingIDPrefix(TrackingIDPrefix)
	return logger, flush, nil
}
