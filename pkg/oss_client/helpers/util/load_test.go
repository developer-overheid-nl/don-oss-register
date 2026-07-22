package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOASVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "openapi.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"info":{"version":"1.2.3"}}`), 0o600))

	version, err := util.LoadOASVersion(path)

	require.NoError(t, err)
	assert.Equal(t, "1.2.3", version)
}

func TestLoadOASVersionReturnsErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := util.LoadOASVersion(filepath.Join(t.TempDir(), "missing.json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not open OAS file")
	})

	t.Run("invalid json", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "openapi.json")
		require.NoError(t, os.WriteFile(path, []byte(`{`), 0o600))

		_, err := util.LoadOASVersion(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse OAS")
	})

	t.Run("missing version", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "openapi.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"info":{}}`), 0o600))

		_, err := util.LoadOASVersion(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version missing from OAS")
	})
}
