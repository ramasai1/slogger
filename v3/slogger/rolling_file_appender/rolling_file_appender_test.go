// Copyright 2013, 2015 MongoDB, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rolling_file_appender

import (
	"github.com/mongodb/slogger/v3/slogger"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const rfaTestLogDir = "log"
const rfaTestLogFilename = "logger_rfa_test.log"
const rfaStderrPath = "stderr.log"
const rfaTestLogPath = rfaTestLogDir + "/" + rfaTestLogFilename
const rfaStdoutLogPath = rfaTestLogDir + "/" + rfaStderrPath

var stderr = os.Stderr

func TestLog(test *testing.T) {
	defer teardown()

	appender, logger := setup(test, 1000, 0, 10, false)
	defer appender.Close()

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is a log message")
	assertCurrentLogContains(test, "This is a log message")

	appender.SetLevel(slogger.WARN)
	logAndWaitAndCheckError(test, logger, slogger.INFO, "This is an info message being fed to a WARN logger")
	assertCurrentLogDoesNotContain(test, "This is an info message being fed to a WARN logger")

	flushAndWaitAndCheckError(test, logger)
}

func TestNoRotation(test *testing.T) {
	defer teardown()

	appender, logger := setup(test, 1000, 0, 10, false)
	defer appender.Close()

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is under 1,000 characters and should not cause a log rotation")
	assertCurrentLogContains(test, "This is under 1,000 characters and should not cause a log rotation")
	flushAndWaitAndCheckError(test, logger)
	assertNumLogFiles(test, 1)
}

func TestNoRotation2(test *testing.T) {
	defer teardown()

	appender, logger := setup(test, -1, 0, 10, false)
	defer appender.Close()

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This should not cause a log rotation")
	flushAndWaitAndCheckError(test, logger)
	assertNumLogFiles(test, 1)
}

func TestOldLogRemoval(test *testing.T) {
	defer teardown()

	appender, logger := setup(test, 10, 0, 2, false)
	defer appender.Close()

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is more than 10 characters and should cause a log rotation")
	flushAndWaitAndCheckError(test, logger)
	assertNumLogFiles(test, 2)

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is more than 10 characters and should cause a log rotation")
	flushAndWaitAndCheckError(test, logger)
	assertNumLogFiles(test, 3)

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is more than 10 characters and should cause a log rotation")
	flushAndWaitAndCheckError(test, logger)
	assertNumLogFiles(test, 3)
}

func TestPreRotation(test *testing.T) {
	createLogDir(test)
	switchStdoutToFile()

	file, err := os.Create(rfaTestLogPath)
	if err != nil {
		test.Fatalf("Failed to create empty logfile: %v", err)
	}

	err = file.Close()
	if err != nil {
		test.Fatalf("Failed to close logfile: %v", err)
	}

	appender, logger := newAppenderAndLogger(test, 1000, 0, 2, true)
	defer appender.Close()
	flushAndWaitAndCheckError(test, logger)
	assertNumLogFiles(test, 2)
}

func TestRotationSizeBased(test *testing.T) {
	defer teardown()

	appender, logger := setup(test, 10, 0, 10, false)
	defer appender.Close()

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is more than 10 characters and should cause a log rotation")
	flushAndWaitAndCheckError(test, logger)

	assertNumLogFiles(test, 2)
}

func TestRotationTimeBased(test *testing.T) {
	defer teardown()

	func() {
		appender, logger := setup(test, -1, time.Second, 10, false)
		defer appender.Close()

		assertNumLogFiles(test, 1)
		time.Sleep(time.Second + 50*time.Millisecond)
		logAndWaitAndCheckError(test, logger, slogger.WARN, "Trigger log rotation 1")
		assertNumLogFiles(test, 2)

		time.Sleep(time.Second + 50*time.Millisecond)
		logAndWaitAndCheckError(test, logger, slogger.WARN, "Trigger log rotation 2")
		assertNumLogFiles(test, 3)
		flushAndWaitAndCheckError(test, logger)
	}()

	// Test that time-based log rotation still works if we recreate
	// the appender.  This forces the state file to be read in
	appender, logger := newAppenderAndLogger(test, -1, time.Second, 10, false)
	defer appender.Close()

	assertNumLogFiles(test, 3)
	time.Sleep(time.Second + 50*time.Millisecond)
	logAndWaitAndCheckError(test, logger, slogger.WARN, "Trigger log rotation 3")
	assertNumLogFiles(test, 4)
	flushAndWaitAndCheckError(test, logger)
}

func TestRotationManual(test *testing.T) {
	defer teardown()
	appender, _ := setup(test, -1, 0, 10, false)
	defer appender.Close()

	assertNumLogFiles(test, 1)

	if err := appender.Rotate(); err != nil {
		test.Fatal("appender.Rotate() returned an error: " + err.Error())
	}
	assertNumLogFiles(test, 2)

	if err := appender.Rotate(); err != nil {
		test.Fatal("appender.Rotate() returned an error: " + err.Error())
	}
	assertNumLogFiles(test, 3)
}

func TestReopen(test *testing.T) {
	defer teardown()

	// simulate manual log rotation via Reopen()

	appender, logger := setup(test, 0, 0, 0, false)
	defer appender.Close()

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is a log message 1")
	flushAndWaitAndCheckError(test, logger)

	assertCurrentLogContains(test, "This is a log message 1")

	rotatedLogPath := rfaTestLogPath + ".rotated"
	if err := os.Rename(rfaTestLogPath, rotatedLogPath); err != nil {
		test.Fatalf("os.Rename() returned an error: %v", err)
	}

	if _, err := os.Stat(rfaTestLogPath); err == nil {
		test.Fatal(rfaTestLogPath + " should not exist after rename")
	}

	assertLogContains(test, rotatedLogPath, "This is a log message 1")

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is a log message 2")
	flushAndWaitAndCheckError(test, logger)

	assertLogContains(test, rotatedLogPath, "This is a log message 2")
	if err := appender.Reopen(); err != nil {
		test.Fatalf("Error reopening log: %v", err)
	}

	assertLogContains(test, rotatedLogPath, "This is a log message 1")
	assertLogContains(test, rotatedLogPath, "This is a log message 2")

	assertCurrentLogDoesNotContain(test, "This is a log message 1")
	assertCurrentLogDoesNotContain(test, "This is a log message 2")

	logAndWaitAndCheckError(test, logger, slogger.WARN, "This is a log message 3")
	flushAndWaitAndCheckError(test, logger)

	assertCurrentLogContains(test, "This is a log message 3")
	assertLogDoesNotContain(test, rotatedLogPath, "This is a log message 3")
}

func TestCompressionOnRotation(test *testing.T) {
	defer teardown()

	appender, logger := setup(test, 10, 0, 10, false)
	appender.compressRotatedLogs = true
	appender.maxUncompressedLogs = 1
	defer appender.Close()

	compressibleMessage := strings.Repeat("This string is easily compressible", 1000)

	logAndWaitAndCheckError(test, logger, slogger.WARN, compressibleMessage)
	flushAndWaitAndCheckError(test, logger)

	checkFiles := func() (compressedLogFiles, sizeCompressedFile int) {
		err := filepath.Walk(rfaTestLogDir, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), ".gz") {
				compressedLogFiles++
				sizeCompressedFile = int(info.Size())
			}
			return err
		})
		if err != nil {
			test.Error(err)
		}
		return
	}

	compressedLogFiles, _ := checkFiles()
	assertNumLogFiles(test, 2)
	if compressedLogFiles != 0 {
		test.Errorf("expected to find no compressed log files")
	}

	logAndWaitAndCheckError(test, logger, slogger.WARN, compressibleMessage)
	flushAndWaitAndCheckError(test, logger)
	compressedLogFiles, sizeCompressedFile := checkFiles()
	assertNumLogFiles(test, 3)
	if compressedLogFiles != 1 {
		test.Errorf("expected to find one compressed log files")
	}
	if sizeCompressedFile >= len(compressibleMessage)/10 {
		test.Errorf("expected the compressed log file size %v to be smaller than the logged bytes %v", sizeCompressedFile, len(compressibleMessage))
	}
}

func assertCurrentLogContains(test *testing.T, expected string) {
	assertLogContains(test, rfaTestLogPath, expected)
}

func assertCurrentLogDoesNotContain(test *testing.T, notExpected string) {
	assertLogDoesNotContain(test, rfaTestLogPath, notExpected)
}

func assertLogContains(test *testing.T, logPath, expected string) {
	actual := readLog(test, logPath)

	if !strings.Contains(actual, expected) {
		test.Errorf("Log %s contains: \n%s\ninstead of\n%s", logPath, actual, expected)
	}
}

func assertLogDoesNotContain(test *testing.T, logPath, notExpected string) {
	actual := readLog(test, logPath)

	if strings.Contains(actual, notExpected) {
		test.Errorf("Log %s should not contain: \n%s", logPath, notExpected)
	}
}

func assertNumLogFiles(test *testing.T, expected_n int) {
	actual_n, err := numLogFiles()
	if err != nil {
		test.Fatal("Could not get numLogFiles")
	}

	if expected_n != actual_n {
		test.Errorf(
			"Expected number of log files to be %d, not %d",
			expected_n,
			actual_n,
		)
	}
}

func createLogDir(test *testing.T) {
	os.RemoveAll(rfaTestLogDir)
	err := os.MkdirAll(rfaTestLogDir, 0777)

	if err != nil {
		test.Fatal("setup() failed to create directory: " + err.Error())
	}
}

func newAppenderAndLogger(test *testing.T, maxFileSize int64, maxDuration time.Duration, maxRotatedLogs int, rotateIfExists bool) (appender *RollingFileAppender, logger *slogger.Logger) {
	builder := NewBuilder(
		rfaTestLogPath,
		maxFileSize,
		maxDuration,
		maxRotatedLogs,
		rotateIfExists,
		func() []string {
			return []string{"This is a header", "more header"}
		},
	)

	builder.minAllowedLogLevel = slogger.INFO
	appender, err := builder.Build()
	if err != nil {
		test.Fatal("NewRollingFileAppender() failed: " + err.Error())
	}

	logger = &slogger.Logger{
		Prefix:    "rfa",
		Appenders: []slogger.Appender{appender},
	}

	return
}

func numLogFiles() (int, error) {
	cwd, err := os.Open(rfaTestLogDir)
	if err != nil {
		return -1, err
	}
	defer cwd.Close()

	var filenames []string
	filenames, err = cwd.Readdirnames(-1)
	if err != nil {
		return -1, err
	}

	visibleFilenames := make([]string, 0, len(filenames)-1)
	for _, filename := range filenames {
		if !strings.HasPrefix(filename, ".") {
			visibleFilenames = append(visibleFilenames, filename)
		}
	}

	// Subtracting 1 to account for stdout.log (which is overridden in the test).
	return len(visibleFilenames) - 1, nil
}

func readStdout(test *testing.T) string {
	bytes, err := ioutil.ReadFile(rfaStdoutLogPath)
	if err != nil {
		test.Fatal("Could not read stdout file")
	}

	return string(bytes)
}

func readLog(test *testing.T, logPath string) string {
	bytes, err := ioutil.ReadFile(logPath)
	if err != nil {
		test.Fatal("Could not read log file")
	}

	return string(bytes)
}

func logAndWaitAndCheckError(test *testing.T, logger *slogger.Logger, level slogger.Level, msgFmt string, args ...any) {
	logger.Logf(level, msgFmt, args...)
	time.Sleep(30 * time.Millisecond)
	assertLogDoesNotContain(test, rfaStdoutLogPath, "Encountered an error while logging.")
}

func flushAndWaitAndCheckError(test *testing.T, logger *slogger.Logger) {
	logger.Flush()
	time.Sleep(50 * time.Millisecond)
	assertLogDoesNotContain(test, rfaStdoutLogPath, "Encountered an error while logging.")
}

func setup(test *testing.T, maxFileSize int64, maxDuration time.Duration, maxRotatedLogs int, rotateIfExists bool) (appender *RollingFileAppender, logger *slogger.Logger) {
	createLogDir(test)
	switchStdoutToFile()

	return newAppenderAndLogger(test, maxFileSize, maxDuration, maxRotatedLogs, rotateIfExists)
}

func switchStdoutToFile() {
	temp, _ := os.Create(rfaStdoutLogPath) // create temp file
	os.Stderr = temp
}

func teardown() {
	os.RemoveAll(rfaTestLogDir)
	os.RemoveAll(rfaStdoutLogPath)
	os.Stderr = stderr
}
