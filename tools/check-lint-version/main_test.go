package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWorkflow(t *testing.T) {
	t.Run("parse invalid workflow file", func(t *testing.T) {
		parsedVersion, err := parseWorkflowVersionFromFile("../../.github/workflows/invalid.yaml")
		assert.Equal(t, "", parsedVersion)
		require.Error(t, err)
	})

	t.Run("parse workflow file with a missing key", func(t *testing.T) {
		parsedVersion, err := parseWorkflowVersionFromFile("./testing/invalid-test.yml")
		assert.Equal(t, "", parsedVersion)
		require.NoError(t, err)
	})

	t.Run("parse an invalid workflow file", func(t *testing.T) {
		parsedVersion, err := parseWorkflowVersionFromFile("./testing/invalid-yaml.yml")
		assert.Equal(t, "", parsedVersion)
		require.Error(t, err)
	})

	t.Run("parse testing workflow file", func(t *testing.T) {
		parsedVersion, err := parseWorkflowVersionFromFile("../../.github/workflows/test-tooling.yml")
		assert.Equal(t, "v1.55.2", parsedVersion)
		require.NoError(t, err)
	})
}

func TestGetCurrentVersion(t *testing.T) {
	t.Run("get current version from system", func(t *testing.T) {
		currentVersion, err := getCurrentVersion()
		assert.Equal(t, "v1.55.2", currentVersion)
		require.NoError(t, err)
	})

	// TODO: test failure to detect current version

	// TODO: test failure to compile regex expression

	// TODO: test failure finding matches
}

func TestIsVersionValid(t *testing.T) {
	t.Run("compare versions - exactly equal to", func(t *testing.T) {
		assert.True(t, true, isVersionValid("v1.54.2", "v1.54.2"))
	})

	t.Run("compare versions - patch version greater (workflow)", func(t *testing.T) {
		assert.True(t, true, isVersionValid("v1.54.3", "v1.54.2"))
	})

	t.Run("compare versions - patch version greater (installed)", func(t *testing.T) {
		assert.True(t, true, isVersionValid("v1.54.2", "v1.54.3"))
	})

	t.Run("compare versions - invalid (installed)", func(t *testing.T) {
		assert.False(t, false, isVersionValid("v1.54.2", "v1.52.2"))
	})

	t.Run("compare versions - invalid (workflow)", func(t *testing.T) {
		assert.False(t, false, isVersionValid("v1.52.2", "v1.54.2"))
	})
}

func TestCompareVersions(t *testing.T) {
	t.Run("Valid comparison", func(t *testing.T) {
		res := compareVersions("../../.github/workflows/test-on-push.yaml")
		assert.Contains(t, res, "Linter version is valid")
	})

	t.Run("Invalid comparison", func(t *testing.T) {
		res := compareVersions("./testing/invalid-test.yml")
		assert.Contains(t, res, "Invalid version")
	})

	// TODO: test function for failure to get the current version using getCurrentVersion()

	t.Run("Invalid path for comparison", func(t *testing.T) {
		res := compareVersions("./testing/invalid-test-incorrect-path.yml")
		assert.Contains(t, res, "Error parsing workflow")
	})
}
