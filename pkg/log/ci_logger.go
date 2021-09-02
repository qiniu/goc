/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package log

import "go.uber.org/zap"

type ciLogger struct {
	logger *zap.Logger
}

func newCiLogger() *ciLogger {
	logger, _ := zap.NewDevelopment()
	// fix: increases the number of caller from always reporting the wrapper code as caller
	logger = logger.WithOptions(zap.AddCallerSkip(2))
	zap.ReplaceGlobals(logger)
	return &ciLogger{
		logger: logger,
	}
}

func (c *ciLogger) StartWait(message string) {

}

func (c *ciLogger) StopWait() {

}

func (c *ciLogger) Sync() {
	c.logger.Sync()
}

func (c *ciLogger) Debugf(format string, args ...interface{}) {
	zap.S().Debugf(format, args...)
}

func (c *ciLogger) Donef(format string, args ...interface{}) {
	zap.S().Infof(format, args...)
}

func (c *ciLogger) Infof(format string, args ...interface{}) {
	zap.S().Infof(format, args...)
}

func (c *ciLogger) Errorf(format string, args ...interface{}) {
	zap.S().Errorf(format, args...)
}

func (c *ciLogger) Warnf(format string, args ...interface{}) {
	zap.S().Warnf(format, args...)
}

func (c *ciLogger) Fatalf(format string, args ...interface{}) {
	zap.S().Fatalf(format, args...)
}
