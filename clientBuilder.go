package eventuate

import loglib "github.com/shopcookeat/eventuate-client-golang/logger"

type ClientBuilderInstance struct {
	overrideCredentials bool
	overrideUrl         bool
	overrideSpace       bool
	overrideStompUrl    bool
	ll                  loglib.LogLevelEnum
	lg                  loglib.Logger
	apiKeyId            string
	apiKeySecret        string
	url                 string
	space               string
	stompUrl            string
	typeHints           typeHintsMap
}

func ClientBuilder() *ClientBuilderInstance {
	return &ClientBuilderInstance{
		false,
		false,
		false,
		false,
		0,
		loglib.NewNilLogger(),
		"",
		"",
		"https://api.eventuate.io",
		"default",
		"https://api.eventuate.io:61614",
		nil}
}

func (bldr *ClientBuilderInstance) WithUrl(serverUrl string) *ClientBuilderInstance {
	bldr.overrideUrl = true
	bldr.url = serverUrl
	return bldr
}
func (bldr *ClientBuilderInstance) WithStompUrl(serverUrl string) *ClientBuilderInstance {
	(*bldr).overrideStompUrl = true
	(*bldr).stompUrl = serverUrl
	return bldr
}
func (bldr *ClientBuilderInstance) WithSpace(space string) *ClientBuilderInstance {
	bldr.overrideSpace = true
	bldr.space = space
	return bldr
}
func (bldr *ClientBuilderInstance) WithCredentials(apiKeyId string,
	apiKeySecret string) *ClientBuilderInstance {
	bldr.overrideCredentials = true
	bldr.apiKeyId = apiKeyId
	bldr.apiKeySecret = apiKeySecret
	return bldr
}

func (bldr *ClientBuilderInstance) WithNewTypeHints() *ClientBuilderInstance {
	bldr.typeHints = NewTypeHintsMap()
	return bldr
}

//func (bldr *ClientBuilderInstance) WithTypeHintPair(name string, typeInstance interface{}) *ClientBuilderInstance {
//
//}

func (bldr *ClientBuilderInstance) WithTypeHintPair(name string, typeInstance interface{}) *ClientBuilderInstance {
	if bldr.typeHints == nil {
		bldr.typeHints = NewTypeHintsMap()
	}
	err := bldr.typeHints.RegisterEventType(name, typeInstance)
	if err != nil {
		panic(err) // we need to panic as it is for the developers to see
	}
	return bldr
}

func (bldr *ClientBuilderInstance) SetLogLevel(level loglib.LogLevelEnum) *ClientBuilderInstance {
	bldr.ll = level
	bldr.lg = loglib.NewLogger(level)
	return bldr
}

func (bldr *ClientBuilderInstance) BuildREST() (*RESTClient, error) {
	var (
		credentials *Credentials
		err         error
	)
	if (*bldr).overrideCredentials {
		credentials, err = NewCredentials(bldr.apiKeyId, bldr.apiKeySecret, bldr.space)
	} else {
		credentials, err = NewCredentialsFromEnvAndSpace(bldr.space)
	}
	if err != nil {
		return nil, err
	}

	result, clientErr := NewRESTClient(credentials, bldr.url)
	if clientErr == nil {
		result.ll = bldr.ll
		result.lg = bldr.lg
	}

	result.typeHints = bldr.typeHints.MakeCopy()

	return result, clientErr
}

func (bldr *ClientBuilderInstance) BuildSTOMP() (*StompClient, error) {
	var (
		credentials *Credentials
		err         error
	)
	if bldr.overrideCredentials {
		credentials, err = NewCredentials(bldr.apiKeyId, bldr.apiKeySecret, bldr.space)
	} else {
		credentials, err = NewCredentialsFromEnvAndSpace(bldr.space)
	}

	if err != nil {
		return nil, err
	}

	result, clientErr := NewStompClient(credentials, bldr.stompUrl)
	if clientErr == nil {
		result.ll = bldr.ll
		result.lg = bldr.lg
	}
	result.typeHints = bldr.typeHints.MakeCopy()

	return result, clientErr
}
