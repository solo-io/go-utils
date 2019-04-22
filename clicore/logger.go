package clicore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CliLoggerKey is the key passed through zap logs that indicates that its value should be written to the console,
// in addition to the full log file.
const CliLoggerKey = "cli"

// BuildCliLogger creates a set of loggers for use in CLI applications.
// - A json-formatted file logger that writes all log messages to the specified filename
// - A human-friendly console logger that writes info and warning messages to stdout
// - A human-friendly console logger that writes info and warning messages to stderr
func BuildCliLogger(pathElements []string, outputModeEnvVar string) *zap.SugaredLogger {
	return buildCliLoggerOptions(pathElements, outputModeEnvVar, nil)
}

//BuildMockedCliLogger is the test-environment counterpart of BuildCliLogger
// It stores log output in buffers that can be inspected by tests.
func BuildMockedCliLogger(pathElements []string, outputModeEnvVar string, mockTargets *MockTargets) *zap.SugaredLogger {
	return buildCliLoggerOptions(pathElements, outputModeEnvVar, mockTargets)
}

func buildCliLoggerOptions(pathElements []string, outputModeEnvVar string, mockTargets *MockTargets) *zap.SugaredLogger {
	verboseMode := os.Getenv(outputModeEnvVar) == "1"
	fileCore := buildCliZapCoreFile(pathElements, verboseMode, mockTargets)
	consoleCores := buildCliZapCoreConsoles(verboseMode, mockTargets)
	allCores := consoleCores
	if fileCore != nil {
		allCores = append(allCores, fileCore)
	}
	core := zapcore.NewTee(allCores...)
	logger := zap.New(core).Sugar()
	return logger
}

//FilePathFromHomeDir is a utility that makes it easier to find the absolute path to a file, given its file path
// elements relative to its home directory.
// pathElementsRelativeToHome is passed as an array to avoid os-specific directory delimiter complications
// example: []string{".config","default.yaml"}
func FilePathFromHomeDir(pathElementsRelativeToHome []string) (string, error) {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	nPathElements := len(pathElementsRelativeToHome)
	nDirElements := nPathElements - 1
	if nDirElements > 0 {
		dirsToMake := append([]string{home}, pathElementsRelativeToHome[:nDirElements]...)
		// if the directory already exists, MkdirAll returns nil
		if err := os.MkdirAll(filepath.Join(dirsToMake...), 0755); err != nil {
			return "", err
		}
	}
	pathElements := append([]string{home}, pathElementsRelativeToHome...)
	return filepath.Join(pathElements...), nil
}

func buildCliZapCoreFile(pathElements []string, verboseMode bool, mockTargets *MockTargets) zapcore.Core {
	// force to verbose mode while running in a test environment
	if mockTargets != nil {
		verboseMode = true
	}
	path, err := FilePathFromHomeDir(pathElements)
	if err != nil {
		if verboseMode {
			// we don't want to return errors just because we cannot write logs to a file
			// users can use the verbose flag to get full output to the console
			fmt.Printf("Could not open log file %s for writing: %v\n", filepath.Join(pathElements...), err)
		}
		return nil
	}

	// if we decide we want to append logs, we can do it this way:
	// (we would be required to first create the file and "rotate" it as it grows)
	//file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, os.ModeAppend)

	// for now, let's just create/overwrite the file each time
	file, err := os.Create(path)
	if err != nil {
		if verboseMode {
			// we don't want to return errors just because we cannot write logs to a file
			// users can use the verbose flag to get full output to the console
			fmt.Printf("Could not open log file %s for writing: %v\n", path, err)
		}
		return nil
	}

	// we want all messages to go to the file log
	passAllMessages := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
	})

	// apply zap's lock and WriteSyncer helpers
	fileDebug := zapcore.Lock(zapcore.AddSync(file))
	if mockTargets != nil {
		fileDebug = zapcore.Lock(mockTargets.FileLog)
	}
	fileLoggerEncoderConfig := defaultEncoderConfig()
	fileEncoder := zapcore.NewJSONEncoder(fileLoggerEncoderConfig)
	fileCore := zapcore.NewCore(fileEncoder, fileDebug, passAllMessages)

	return fileCore
}

func buildCliZapCoreConsoles(verboseMode bool, mockTargets *MockTargets) []zapcore.Core {
	// force to verbose mode while running in a test environment
	if mockTargets != nil {
		verboseMode = true
	}

	// define error filter levels
	errorMessages := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	stdOutMessages := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.DebugLevel && lvl < zapcore.ErrorLevel
	})
	stdOutMessagesVerbose := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	// add locks for safe concurrency
	consoleInfo := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)
	if mockTargets != nil {
		consoleInfo = zapcore.Lock(mockTargets.Stdout)
		consoleErrors = zapcore.Lock(mockTargets.Stderr)
	}

	consoleLoggerEncoderConfig := defaultEncoderConfig()

	// minimize the noise for non-verbose mode
	if !verboseMode {
		consoleLoggerEncoderConfig.EncodeTime = nil
		consoleLoggerEncoderConfig.LevelKey = ""
		consoleLoggerEncoderConfig.NameKey = ""
	}
	consoleEncoder := NewCliEncoder(CliLoggerKey)

	consoleStdoutCore := zapcore.NewCore(consoleEncoder, consoleInfo, stdOutMessages)
	if verboseMode {
		consoleStdoutCore = zapcore.NewCore(consoleEncoder, consoleInfo, stdOutMessagesVerbose)
	}
	consoleErrCore := zapcore.NewCore(consoleEncoder, consoleErrors, errorMessages)
	return []zapcore.Core{consoleStdoutCore, consoleErrCore}
}

func defaultEncoderConfig() zapcore.EncoderConfig {
	fileLoggerEncoderConfig := zap.NewProductionEncoderConfig()
	fileLoggerEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return fileLoggerEncoderConfig
}
