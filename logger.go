package main

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Global logger variables
var (
	logFile      *os.File
	logFileMutex sync.Mutex
)

// initLogger initializes the global file logger.
func initLogger() {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	var err error
	// Ensure the /logs directory exists if running locally without Docker volume mount
	// os.MkdirAll("/logs", 0755) // Uncomment if needed

	logFile, err = os.OpenFile("/logs/todo-app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Fallback to stderr if file logging fails
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: false})
		log.Error().Err(err).Msg("Failed to open log file, falling back to stderr")
		return
	}

	// Use ConsoleWriter for potentially colored output if TTY, disable color for file.
	// Or keep color if the log file viewer supports ANSI codes.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile, NoColor: true}) // Typically disable color for files
	log.Info().Msg("Logger initialized to file /logs/todo-app.log")
}

// reopenLogFile periodically closes and reopens the log file for rotation/external management.
func reopenLogFile() {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if logFile == nil {
		log.Error().Msg("Log file is nil, cannot reopen")
		// Attempt to re-initialize
		initLogger() // Be careful about potential recursion if initLogger fails repeatedly
		return
	}

	// Close the current log file
	if err := logFile.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close log file during reopen")
		// Continue trying to reopen anyway
	}

	// Reopen the log file
	var err error
	logFile, err = os.OpenFile("/logs/todo-app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reopen log file")
		// Consider falling back to stderr or another strategy
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: false})
		logFile = nil // Mark logFile as nil since it failed
		return
	}

	// Update the logger to use the new file handle
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile, NoColor: true}) // Keep color setting consistent
	log.Info().Msg("Log file reopened")
}

// closeLogFile safely closes the global log file.
func closeLogFile() {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()
	if logFile != nil {
		log.Info().Msg("Closing log file.")
		logFile.Sync() // Attempt to flush buffer
		err := logFile.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing log file")
		}
		logFile = nil
	}
}

// startLogRotation starts a goroutine to periodically reopen the log file.
func startLogRotation() {
	go func() {
		// Adjust interval as needed, e.g., for daily rotation link with signal handling
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			log.Info().Msg("Scheduled log file reopen triggered.")
			reopenLogFile()
		}
	}()
}
