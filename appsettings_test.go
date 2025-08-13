package appsettings_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/prutswonder/go-appsettings" // Adjust the import path as necessary
	"github.com/stretchr/testify/assert"
)

type TestSettings struct {
	Global struct {
		Log struct {
			Level string `json:"msg-level"`
		}
	}
	Cors struct {
		Origins []string
	}
	Custom struct {
		Service struct {
			Name string
		}
		Enabled bool
	}
	Google struct {
		App struct {
			Credentials string
		}
	}
}

func (s *TestSettings) Validate(settings any) error {
	errs := error(nil)
	if s.Global.Log.Level == "" {
		errs = errors.Join(errs, fmt.Errorf("Global.Log.Level is required"))
	}
	if s.Google.App.Credentials == "" {
		errs = errors.Join(errs, fmt.Errorf("Google.App.Credentials is required"))
	}
	return errs
}

func TestAppSettings(t *testing.T) {
	settings := &TestSettings{}

	// By default this repository does not have an appsettings.json file, so this should fail.
	err := appsettings.ReadSettingsFromFileAndEnv(settings, nil)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "open appsettings file")
	assert.ErrorContains(t, err, "open appsettings.json")

	// Save the current environment variables to restore later, just in case they are set.
	logMinFilter := os.Getenv("GLOBAL_LOG_LEVEL")
	credentials := os.Getenv("GOOGLE_APP_CREDENTIALS")

	// Create a faulty appsettings.json file.
	notJsonContent := `{
			"global": {
				"log": {
					"msg-level" "Debug"
				}}
			},
			cors": {
				"origins": ["*"]
			]
		`
	err = os.WriteFile("appsettings.json", []byte(notJsonContent), 0644)
	assert.NoError(t, err)

	// Although appsettings.json exists now, it is not valid JSON, so this should fail.
	err = appsettings.ReadSettingsFromFileAndEnv(settings, nil)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "unmarshal appsettings file")

	// Create an appsettings.json file with some default values.
	jsonContent := `{
			"global": {
				"log": {
					"msg-level": "Debug"
				}
			},
			"cors": {
				"origins": ["*"]
			}
		}`
	err = os.WriteFile("appsettings.json", []byte(jsonContent), 0644)
	assert.NoError(t, err)

	// Restore the environment variables and delete appsettings.json after the test.
	defer func() {
		if err = os.Remove("appsettings.json"); err != nil {
			t.Errorf("Failed to remove appsettings.json: %v", err)
		}
		if err := os.Setenv("GLOBAL_LOG_LEVEL", logMinFilter); err != nil {
			t.Errorf("Failed to restore GLOBAL_LOG_LEVEL: %v", err)
		}
		if err := os.Setenv("GOOGLE_APP_CREDENTIALS", credentials); err != nil {
			t.Errorf("Failed to restore GOOGLE_APP_CREDENTIALS: %v", err)
		}
	}()

	// Clear the environment variables to ensure they do not interfere with the test.
	if err := os.Setenv("GLOBAL_LOG_LEVEL", ""); err != nil {
		t.Errorf("Failed to write GLOBAL_LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("GOOGLE_APP_CREDENTIALS", ""); err != nil {
		t.Errorf("Failed to write GOOGLE_APP_CREDENTIALS: %v", err)
	}

	// Now that appsettings.json exists, this should succeed without validation.
	err = appsettings.ReadSettingsFromFileAndEnv(settings, nil)
	assert.NoError(t, err)

	// With validation, this should fail because GOOGLE_APP_CREDENTIALS is not set.
	// LOG_MINFILTER is optional, so it should not cause an error.
	err = appsettings.ReadSettingsFromFileAndEnv(settings, settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "validate settings")
	assert.ErrorContains(t, err, "Google.App.Credentials")
	assert.NotContains(t, err.Error(), "Global.Log.Level")

	if err := os.Setenv("GOOGLE_APP_CREDENTIALS", "something"); err != nil {
		t.Errorf("Failed to write GOOGLE_APP_CREDENTIALS: %v", err)
	}

	// Now that appsettings.json and GOOGLE_APP_CREDENTIALS exist, this should succeed.
	err = appsettings.ReadSettingsFromFileAndEnv(settings, settings)
	assert.NoError(t, err)
	assert.Equal(t, "Debug", settings.Global.Log.Level)
	assert.Equal(t, []string{"*"}, settings.Cors.Origins)
	assert.False(t, settings.Custom.Enabled)
	assert.Equal(t, "something", settings.Google.App.Credentials)
}
