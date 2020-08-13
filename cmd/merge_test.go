/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

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

package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// use a variable to record if the tested function has failed
// so unittest can only be executed sequential
var fatal = false
var fatalStr string

type TestLogHook struct{}

func (h *TestLogHook) Levels() []log.Level {
	return []log.Level{log.FatalLevel}
}

func (h *TestLogHook) Fire(e *log.Entry) error {
	fatalStr = e.Message
	return nil
}

func TestMain(m *testing.M) {
	// setup
	originalExitFunc := log.StandardLogger().ExitFunc
	defer func() {
		log.StandardLogger().ExitFunc = originalExitFunc
		log.StandardLogger().Hooks = make(log.LevelHooks)
	}()

	// replace exit function, so log.Fatal wont exit
	log.StandardLogger().ExitFunc = func(int) { fatal = true }
	// add hook, so fatal string will be recorded
	log.StandardLogger().Hooks.Add(&TestLogHook{})

	code := m.Run()

	os.Exit(code)
}

func TestMergeNormalProfiles(t *testing.T) {
	profileA := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/a.voc")
	profileB := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/b.voc")
	mergeprofile := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/merge.cov")

	runMerge([]string{profileA, profileB}, mergeprofile)

	contents, err := ioutil.ReadFile(mergeprofile)
	assert.NoError(t, err)
	assert.Contains(t, string(contents), "qiniu.com/kodo/apiserver/server/main.go:32.49,33.13 1 60")
	assert.Contains(t, string(contents), "qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 2")
	assert.Equal(t, fatal, false)
}

// test with no profiles
func TestMergeWithNoProfiles(t *testing.T) {
	mergeprofile := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/merge.cov")

	// clear fatal string in setup
	fatalStr = ""
	fatal = false

	runMerge([]string{}, mergeprofile)

	// there is fatal
	assert.Equal(t, fatal, true)
	assert.Equal(t, fatalStr, "Expected at least one coverage file.")
}

// pass a non-existed profile to runMerge
func TestWithWrongProfileName(t *testing.T) {
	profileA := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/notexist.voc")
	mergeprofile := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/merge.cov")

	// clear fatal string in setup
	fatalStr = ""
	fatal = false

	runMerge([]string{profileA}, mergeprofile)

	// there is fatal
	assert.Equal(t, fatal, true)
	assert.Contains(t, fatalStr, "failed to open")
}

// merge two different modes' profiles should fail
func TestMergeTwoDifferentModeProfile(t *testing.T) {
	profileA := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/a.voc")
	profileSet := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/setmode.voc")
	mergeprofile := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/merge.cov")

	// clear fatal string in setup
	fatalStr = ""
	fatal = false

	runMerge([]string{profileA, profileSet}, mergeprofile)

	// there is fatal
	assert.Equal(t, fatal, true)
	assert.Contains(t, fatalStr, "mode for qiniu.com/kodo/apiserver/server/main.go mismatches")
}

// merge two overlaped profiles should fail
func TestMergeTwoOverLapProfile(t *testing.T) {
	profileA := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/a.voc")
	profileOverlap := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/overlap.voc")
	mergeprofile := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/merge.cov")

	// clear fatal string in setup
	fatalStr = ""
	fatal = false

	runMerge([]string{profileA, profileOverlap}, mergeprofile)

	// there is fatal
	assert.Equal(t, fatal, true)
	assert.Contains(t, fatalStr, "coverage block mismatch")
}

// merge empty file should fail
func TestMergeEmptyProfile(t *testing.T) {
	profileA := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/empty.voc")
	mergeprofile := filepath.Join(baseDir, "../tests/samples/merge_profile_samples/merge.cov")

	// clear fatal string in setup
	fatalStr = ""
	fatal = false

	runMerge([]string{profileA}, mergeprofile)

	// there is fatal
	assert.Equal(t, fatal, true)
	assert.Contains(t, fatalStr, "failed to dump profile")
}
