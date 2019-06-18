package test

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CopyTestFile copies the contents of a designated file to a temporary file and returns
// the filepath of the temporary file for cleanup.
// This function does NOT clean up created temporary files.
// Borrows from: https://opensource.com/article/18/6/copying-files-go
func CopyTestFile(t *testing.T, dstDir, dstPrefix, src string) string {
	t.Helper()

	// Create temporary destination file.
	destination, err := ioutil.TempFile(dstDir, dstPrefix)
	require.NoError(t, err)
	defer destination.Close()

	sourceFileStat, err := os.Stat(src)
	require.NoError(t, err)
	require.True(t, sourceFileStat.Mode().IsRegular())

	source, err := os.Open(filepath.Clean(src))
	require.NoError(t, err)
	defer source.Close()

	// Copy file from source to destination.
	_, err = io.Copy(destination, source)
	require.NoError(t, err)

	return destination.Name()
}
