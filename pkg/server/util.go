package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/tools/cover"
)

func convertProfile(p []byte) ([]*cover.Profile, error) {
	// Annoyingly, ParseProfiles only accepts a filename, so we have to write the bytes to disk
	// so it can read them back.
	// We could probably also just give it /dev/stdin, but that'll break on Windows.
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file, err: %v", err)
	}
	defer tf.Close()
	defer os.Remove(tf.Name())
	if _, err := io.Copy(tf, bytes.NewReader(p)); err != nil {
		return nil, fmt.Errorf("failed to copy data to temp file, err: %v", err)
	}

	return cover.ParseProfiles(tf.Name())
}
