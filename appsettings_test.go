package appsettings_test

import (
	"errors"
	"fmt"
	"io"
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

type TestReader struct {
	HasReadError  bool
	HasCloseError bool
}

func (r *TestReader) Read(p []byte) (n int, err error) {
	if r.HasReadError {
		return 0, errors.New("read error")
	}
	return 0, io.EOF
}
func (r *TestReader) Close() error {
	if r.HasCloseError {
		return errors.New("close error")
	}
	return nil
}

type TestUpdater struct {
	LogLevel          string
	GoogleCredentials string
	UpdateError       error
}

func (u *TestUpdater) Update(settings any) error {
	s, _ := settings.(*TestSettings)

	if u.LogLevel != "" {
		s.Global.Log.Level = u.LogLevel
	}
	if u.GoogleCredentials != "" {
		s.Google.App.Credentials = u.GoogleCredentials
	}
	return u.UpdateError
}

func TestAppSettings(t *testing.T) {
	settings := &TestSettings{}

	// A nil AppSettings instance is not allowed.
	sut := (*appsettings.AppSettings)(nil)
	err := sut.Read(settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, appsettings.ErrAppSettingsNil.Error())

	// A nil reader is not allowed.
	sut = &appsettings.AppSettings{}
	err = sut.Read(settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, appsettings.ErrReaderNil.Error())

	// Faulty reader is accepted at instantiation.
	reader := TestReader{HasReadError: true}
	sut, err = appsettings.NewAppSettings(&reader)
	assert.NoError(t, err)

	// Reading settings with a nil parameter should fail.
	err = sut.Read(nil)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, appsettings.ErrSettingsParamNil.Error())

	// Reading settings should fail with the faulty reader.
	err = sut.Read(settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, errors.New("read error").Error())

	reader = TestReader{HasCloseError: true}

	// Reading settings with a reader that fails to close should fail.
	err = sut.Read(settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, errors.New("close error").Error())

	// By default this repository does not have an appsettings.json file, so this should fail.
	_, err = appsettings.NewAppSettings(nil)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, appsettings.ErrOpeningFile.Error())
	assert.ErrorContains(t, err, "open appsettings.json")

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

	// Instantiating AppSettings should succeed, even with a faulty appsettings.json.
	sut, err = appsettings.NewAppSettings(nil)
	assert.NoError(t, err)
	assert.NotNil(t, sut)

	// Although appsettings.json exists now, it is not valid JSON, so this should fail.
	err = sut.Read(settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, appsettings.ErrUnmarshalingFile.Error())

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
	}()

	updater := TestUpdater{}
	sut, err = appsettings.NewAppSettings(nil)
	sut = sut.WithUpdater(&updater)

	// Now that appsettings.json exists, this should succeed without validation.
	err = sut.Read(settings)
	assert.NoError(t, err)

	sut, err = appsettings.NewAppSettings(nil)
	sut = sut.WithUpdater(&updater)
	sut = sut.WithValidator(settings)

	// With validation, this should fail because Google.App.Credentials is not set.
	// Global.Log.Level is set in the JSON file, so it should not cause an error.
	err = sut.Read(settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "validate settings")
	assert.ErrorContains(t, err, "Google.App.Credentials")
	assert.NotContains(t, err.Error(), "Global.Log.Level")

	updater.GoogleCredentials = "something"
	sut, err = appsettings.NewAppSettings(nil)
	sut = sut.WithUpdater(&updater)

	// Now that Google.App.Credentials exists, this should succeed.
	err = sut.Read(settings)
	assert.NoError(t, err)
	assert.Equal(t, "Debug", settings.Global.Log.Level)
	assert.Equal(t, []string{"*"}, settings.Cors.Origins)
	assert.False(t, settings.Custom.Enabled)
	assert.Equal(t, "something", settings.Google.App.Credentials)

	updater.UpdateError = errors.New("updater error")
	sut, err = appsettings.NewAppSettings(nil)
	sut = sut.WithUpdater(&updater)
	sut = sut.WithValidator(settings)

	// This should fail because the updater returns an error.
	err = sut.Read(settings)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "updater error")
}
