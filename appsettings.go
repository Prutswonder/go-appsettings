package appsettings

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type (
	// EnvUpdater is an interface to update settings from other sources, for example environment variables.
	EnvUpdater interface {
		Update(settings any) error
	}

	// Validator is an interface to validate settings after reading them.
	Validator interface {
		Validate(settings any) error
	}

	// Composer is a struct to read application settings from multiple sources and validate them if needed.
	Composer struct {
		jsonReader io.ReadCloser
		updater    EnvUpdater
		validator  Validator
	}
)

var (
	ErrAppSettingsNil   = errors.New("appsettings instance is nil")
	ErrSettingsParamNil = errors.New("settings parameter is nil")
	ErrReaderNil        = errors.New("no settings reader provided")

	ErrOpeningFile      = errors.New("failed to open appsettings file")
	ErrClosingFile      = errors.New("failed to close appsettings file")
	ErrReadingFile      = errors.New("failed to read appsettings file")
	ErrUnmarshalingFile = errors.New("failed to unmarshal appsettings file")

	ErrUpdateSettings   = errors.New("failed to update settings with env vars")
	ErrValidateSettings = errors.New("failed to validate settings")

	DefaultAppSettingsFile = "appsettings.json"
)

// NewComposer creates a new Composer instance.
// It will use the default "appsettings.json" file reader.
func NewComposer() (*Composer, error) {
	return NewComposerWithReader(nil)
}

// NewComposerWithReader creates a new Composer instance.
// If the provided jsonReadCloser is nil, it will use the default "appsettings.json" file reader.
// If the default JSON file cannot be found or opened, it returns an error.
func NewComposerWithReader(jsonReadCloser io.ReadCloser) (*Composer, error) {
	if jsonReadCloser == nil {
		if f, err := os.Open(DefaultAppSettingsFile); err != nil {
			return nil, errors.Join(ErrOpeningFile, err)
		} else {
			jsonReadCloser = f
		}
	}
	as := &Composer{
		jsonReader: jsonReadCloser,
		updater:    nil,
		validator:  nil,
	}
	return as, nil
}

// WithUpdater sets the updater anfor the AppSettings instance.
func (as *Composer) WithUpdater(updater EnvUpdater) *Composer {
	as.updater = updater
	return as
}

// WithValidator sets the validator for the AppSettings instance.
func (as *Composer) WithValidator(validator Validator) *Composer {
	as.validator = validator
	return as
}

// Read reads the settings from multiple sources and validates them if a validator is provided.
func (as *Composer) Read(settings any) (err error) {
	// Step 0: Basic validation
	if as == nil {
		return ErrAppSettingsNil
	}
	if settings == nil {
		return ErrSettingsParamNil
	}
	if as.jsonReader == nil {
		return ErrReaderNil
	}

	// Step 1: Read settings from file and close it after reading
	{
		defer func() {
			if closeErr := as.jsonReader.Close(); closeErr != nil {
				err = errors.Join(ErrClosingFile, closeErr)
			}
		}()
		data, err := io.ReadAll(as.jsonReader)

		if err != nil {
			return errors.Join(ErrReadingFile, err)
		}

		if err = json.Unmarshal(data, settings); err != nil {
			return errors.Join(ErrUnmarshalingFile, err)
		}
	}

	// Step 2: Override with environment variables, in case updater is provided
	if as.updater != nil {
		if err := as.updater.Update(settings); err != nil {
			return errors.Join(ErrUpdateSettings, err)
		}
	}
	// if err := envconfig.InitWithOptions(settings, envconfig.Options{AllOptional: true}); err != nil {
	// 	return errors.Join(fmt.Errorf("failed to update settings with env vars"), err)
	// }

	//Step 3: Validate settings in case a validator is provided
	if as.validator == nil {
		return nil
	}
	if errs := as.validator.Validate(settings); errs != nil {
		return errors.Join(ErrValidateSettings, errs)
	}

	// All good, return nil
	return nil
}
