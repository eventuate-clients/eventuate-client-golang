package eventuate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty"

	"crypto/tls"
	loglib "github.com/eventuate-clients/eventuate-client-golang/logger"
)

type RESTClient struct {
	Credentials *Credentials
	Url         *url.URL
	ll          loglib.LogLevelEnum
	lg          loglib.Logger
	typeHints   typeHintsMap
	resty       *resty.Client
}

func NewRESTClient(credentials *Credentials, serverUrl string) (*RESTClient, error) {

	if credentials == nil {
		return nil, AppError("NewRESTClient: Credentials not provided")
	}

	if len(serverUrl) == 0 {
		return nil, AppError("NewRESTClient: url not provided")
	}

	storeServerUrl, urlErr := url.Parse(serverUrl)
	if urlErr != nil {
		return nil, urlErr
	}

	var rest *RESTClient

	retryOnUnavailability := resty.RetryConditionFunc(func(resp *resty.Response) (bool, error) {

		status := resp.StatusCode()
		switch status {
		case http.StatusOK:
			return false, nil

		case http.StatusServiceUnavailable:
			fallthrough
		case http.StatusRequestTimeout:
			return true, nil
		}
		return false, rest.handleNon200Code(status, storeServerUrl, resp.Request.Body, resp.Body())
	})

	restyClient := resty.
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(credentials.apiKeyId, credentials.apiKeySecret).
		AddRetryCondition(retryOnUnavailability)

	rest = &RESTClient{
		Credentials: credentials,
		Url:         storeServerUrl,
		ll:          loglib.Silent,
		lg:          loglib.NewNilLogger(),
		resty:       restyClient,
		typeHints:   NewTypeHintsMap()}

	return rest, nil
}

func (rest *RESTClient) SetLogLevel(level loglib.LogLevelEnum) {
	rest.lg = loglib.NewLogger(level)
	rest.ll = level
}

func (rest *RESTClient) SetInsecureSkipVerify(flag bool) {
	rest.resty.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: flag})
}

func (rest *RESTClient) Find(
	aggregateType string,
	entityId Int128,
	findOptions *AggregateCrudFindOptions) (*LoadedEvents, error) {

	rest.lg.Printf("Find(type: %s, entityId: %s)", aggregateType, entityId)

	query := url.Values{}
	if findOptions != nil {
		if findOptions.TriggeringEvent != nil {
			query["triggeringEventToken"] = []string{findOptions.TriggeringEvent.String()}
		}
	}

	reqUrl, parseErr := rest.Url.Parse(makeGetUrl(rest.Credentials.Space, aggregateType, entityId.String(), query.Encode()))
	if parseErr != nil {
		rest.lg.Println(parseErr)
		return nil, parseErr
	}

	resp, respErr := rest.resty.R().Get(reqUrl.String())
	if respErr != nil {
		rest.lg.Println(respErr)
		return nil, respErr
	}

	status := resp.StatusCode()
	body := resp.Body()

	rest.lg.Printf("\nFIND:\nURL: %s\nHEADERS: %v\nRESPONSE: %v\n",
		reqUrl, resp.Request.Header, string(body))

	if status == http.StatusOK {
		return rest.handleGetResponse(body)
	}

	return nil, rest.handleNon200Code(status, reqUrl, resp.Request.Body, body)
}

func (rest *RESTClient) Save(aggregateType string,
	events []EventTypeAndData,
	saveOptions *AggregateCrudSaveOptions) (*EntityIdVersionAndEventIds, error) {

	rest.lg.Printf("Save(type: %s, events: %s)", aggregateType, events)

	jsonPayload := make(map[string]interface{})
	jsonPayload["entityTypeName"] = aggregateType
	jsonPayload["events"] = events
	if saveOptions != nil {
		if !saveOptions.EntityId.IsNil() {
			jsonPayload["entityId"] = saveOptions.EntityId
		}
	}

	reqUrl, parseErr := rest.Url.Parse(makeNsUrl(rest.Credentials.Space))
	if parseErr != nil {
		rest.lg.Println(parseErr)
		return nil, parseErr
	}

	reqJson, reqJsonErr := json.Marshal(jsonPayload)
	if reqJsonErr != nil {
		rest.lg.Println(reqJsonErr)
		return nil, reqJsonErr
	}
	reqJsonTxt := string(reqJson)

	resp, respErr := rest.resty.R().SetBody(jsonPayload).Post(reqUrl.String())
	if respErr != nil {
		rest.lg.Println(respErr)
		return nil, respErr
	}

	status := resp.StatusCode()
	body := resp.Body()

	rest.lg.Printf("\nSAVE:\nURL: %s\nBODY: %v\n(%v) \nHEADERS: %v\nRESPONSE: %v\n",
		reqUrl, reqJsonTxt, jsonPayload, resp.Request.Header, string(body))

	if status == http.StatusOK {
		return rest.handleCreateResponse(body)
	}

	return nil, rest.handleNon200Code(status, reqUrl, resp.Request.Body, body)
}

func (rest *RESTClient) Update(
	aggregateIdAndType EntityIdAndType,
	entityVersion Int128,
	events []EventTypeAndData,
	updateOptions *AggregateCrudUpdateOptions) (*EntityIdVersionAndEventIds, error) {

	rest.lg.Printf("Update(type: %s, events: %s)", aggregateIdAndType, events)

	jsonPayload := make(map[string]interface{})
	jsonPayload["entityVersion"] = entityVersion
	jsonPayload["events"] = events
	if updateOptions != nil {
		if updateOptions.TriggeringEvent != nil {
			jsonPayload["triggeringEventToken"] = *updateOptions.TriggeringEvent
		}
	}

	reqUrl, parseErr := rest.Url.Parse(makeUpdateUrl(rest.Credentials.Space, aggregateIdAndType))
	if parseErr != nil {
		rest.lg.Println(parseErr)
		return nil, parseErr
	}

	reqJson, reqJsonErr := json.Marshal(jsonPayload)
	if reqJsonErr != nil {
		rest.lg.Println(reqJsonErr)
		return nil, reqJsonErr
	}
	reqJsonTxt := string(reqJson)

	resp, respErr := rest.resty.R().SetBody(jsonPayload).Post(reqUrl.String())
	if respErr != nil {
		rest.lg.Println(respErr)
		return nil, respErr
	}

	status := resp.StatusCode()
	body := resp.Body()

	rest.lg.Printf("\nUPDATE:\nURL: %s\nBODY: %v\n(%v) \nHEADERS: %v\nRESPONSE: %v\n",
		reqUrl, reqJsonTxt, jsonPayload, resp.Request.Header, string(body))

	if status == http.StatusOK {
		return rest.handleUpdateResponse(body)
	}

	return nil, rest.handleNon200Code(status, reqUrl, resp.Request.Body, body)
}

func makeNsUrl(space string) string {
	if len(space) == 0 {
		space = "default"
	}
	return fmt.Sprintf("/entity/%s", space)
}

func makeUpdateUrl(space string, idAndType EntityIdAndType) string {
	return fmt.Sprintf("%s/%s/%s", makeNsUrl(space), idAndType.EntityType, idAndType.EntityId)
}

func makeGetUrl(space, aggType, entityId, query string) string {
	return fmt.Sprintf("%s/%s/%s?%s", makeNsUrl(space), aggType, entityId, query)
}

func (rest *RESTClient) handleNon200Code(code int, reqUrl *url.URL, reqBody interface{}, respBody []byte) error {

	var conflict string
	if code == http.StatusConflict {

		var (
			jsMap map[string]interface{}
		)
		jsMap = make(map[string]interface{})
		jsonErr := json.Unmarshal(respBody, &jsMap)
		if jsonErr == nil {
			errorProp, isString := jsMap["Error"].(string)
			if isString {
				conflict = errorProp
			}
		}
	}

	return RestError(code, conflict, "URL: %s\nRequest: %#v\nResponse: %v", reqUrl, reqBody, string(respBody))
}

func (rest *RESTClient) handleGetResponse(respBody []byte) (*LoadedEvents, error) {
	var response GetResponse

	err := json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}

	tmp := LoadedEvents(response)

	if rest.ll != loglib.Silent {
		o, _ := json.Marshal(tmp)
		rest.lg.Printf("handleGetResponse() result: <LoadedEvents: %v />\n", string(o))
	}

	return &tmp, nil
}

func (rest *RESTClient) handleCreateResponse(respBody []byte) (*EntityIdVersionAndEventIds, error) {
	var response CreateResponse

	err := json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}

	tmp := EntityIdVersionAndEventIds(response)

	if rest.ll != loglib.Silent {
		o, _ := json.Marshal(tmp)
		rest.lg.Printf("handleCreateResponse() result: <EntityIdVersionAndEventIds: %v />\n", string(o))
	}

	return &tmp, nil
}

func (rest *RESTClient) handleUpdateResponse(respBody []byte) (*EntityIdVersionAndEventIds, error) {
	var response UpdateResponse

	err := json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}

	tmp := EntityIdVersionAndEventIds(response)

	if rest.ll != loglib.Silent {
		o, _ := json.Marshal(tmp)
		rest.lg.Printf("handleUpdateResponse() result: <EntityIdVersionAndEventIds: %v />\n", string(o))
	}

	return &tmp, nil
}
