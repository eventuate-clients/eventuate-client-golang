package eventuate

import (
	"os"
	"strings"
)

type Credentials struct {
	apiKeyId     string
	apiKeySecret string
	Space        string
}

func NewCredentials(
	apiKeyId string,
	apiKeySecret string,
	space string) (*Credentials, error) {

	var (
		missing []string = make([]string, 0)
	)

	switch {
	case len(apiKeyId) == 0:
		{
			missing = append(missing, "apiKeyId")
		}
		fallthrough

	case len(apiKeySecret) == 0:
		{
			missing = append(missing, "apiKeySecret")
			return nil, AppError("NewCredentials: parameters missing: %s",
				strings.Join(missing, ", "))

		}
	case len(space) == 0:
		{
			space = "default"
		}
	}

	return &Credentials{
		apiKeyId,
		apiKeySecret,
		space}, nil
}

func NewCredentialsFromEnv() (*Credentials, error) {
	return NewCredentialsFromEnvAndSpace("default")
}

func NewCredentialsFromEnvAndSpace(space string) (*Credentials, error) {
	return NewCredentials(
		os.Getenv("EVENTUATE_API_KEY_ID"),
		os.Getenv("EVENTUATE_API_KEY_SECRET"),
		space)
}
