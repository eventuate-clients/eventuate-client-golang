package eventuate

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/gmallard/stompngo"
	loglib "github.com/eventuate-clients/eventuate-client-golang/logger"
)

type StompClient struct {
	credentials     *Credentials
	Url             *url.URL
	ll              loglib.LogLevelEnum
	lg              loglib.Logger
	stompConnection *stompngo.Connection
	typeHints       typeHintsMap
}

func (stomp *StompClient) RegisterEventType(name string, typeInstance interface{}) error {
	return stomp.typeHints.RegisterEventType(name, typeInstance)
}

func NewStompClient(credentials *Credentials, serverUrl string) (*StompClient, error) {
	if credentials == nil {
		return nil, AppError("NewStompClient: Credentials not provided")
	}
	if serverUrl == "" {
		return nil, AppError("NewStompClient: url not provided")
	}

	stompServerUrl, urlErr := url.Parse(serverUrl)
	if urlErr != nil {
		return nil, urlErr
	}
	return &StompClient{
		credentials: credentials,
		Url:         stompServerUrl,
		ll:          loglib.Silent,
		lg:          loglib.NewNilLogger(),
		typeHints:   NewTypeHintsMap()}, nil
}

func (stomp *StompClient) SubscribeAndDispatch(
	subscriberId string,
	eventHandlers *EventResultHandlerMap,
	subscriberOptions *SubscriberOptions,
	useSwimlane bool) (*DispatchingSubscription, error) {

	return newSubscriptionManager(stomp).Subscribe(subscriberId, eventHandlers, subscriberOptions, useSwimlane)
}

func (stomp *StompClient) Subscribe(
	subscriberId string,
	aggregatesAndEvents map[string][]string,
	subscriberOptions *SubscriberOptions,
	handler *EventResultHandler) (*Subscription, error) {

	if stomp.stompConnection == nil {
		_, err := stomp.makeStompConnection()
		if err != nil {
			return nil, err
		}
	}

	// #1
	uid := stompngo.Uuid()

	// #2.1/3
	sr := &SubscriptionRequest{
		EntityTypesAndEvents: aggregatesAndEvents,
		SubscriberID:         subscriberId,
		Space:                stomp.credentials.Space,
	}

	// #2.2/3
	dhJson, err := json.Marshal(sr)

	if err != nil {
		return nil, err
	}

	// #2.3/3
	destinationHeader := string(dhJson)

	headers := stompngo.Headers{
		"id", fmt.Sprintf("subscription-id-%v", uid),
		"destination", destinationHeader,
		//"heart-beat", "0,10000",
		"ack", stompngo.AckModeClientIndividual}

	receiveChan, subscrErr := stomp.stompConnection.Subscribe(headers)
	if subscrErr != nil {
		return nil, subscrErr
	}

	subscription := newSubscription(uid, stomp.stompConnection, receiveChan, handler)
	subscription.SetLogLevel(stomp.ll)
	subscription.unsubscribeFn = func() error {
		return stomp.stompConnection.Unsubscribe(headers)
	}

	return subscription, nil
}

func (stomp *StompClient) makeStompConnection() (*stompngo.Connection, error) {

	var (
		conn    *stompngo.Connection
		err     error
		netConn net.Conn
	)

	netConn, netErr := stomp.makeTcpConnection()

	if netErr != nil {
		return nil, netErr
	}

	headers := stompngo.Headers{
		stompngo.HK_ACCEPT_VERSION, "1.2",
		stompngo.HK_HOST, stomp.getHost(),
		stompngo.HK_LOGIN, stomp.credentials.apiKeyId,
		stompngo.HK_PASSCODE, stomp.credentials.apiKeySecret,
		//stompngo.HK_SUPPRESS_CL, "1",
		stompngo.HK_HEART_BEAT, "0,10000", // KEEP
	}

	stompConn, stompErr := stompngo.Connect(netConn, headers)

	if stompErr != nil {
		return nil, stompErr
	}

	// Show connect response
	stomp.lg.Printf("connsess:%s common_connect_response connresp:%v\n",
		stompConn.Session(),
		stompConn.ConnectResponse)
	// Heartbeat Data
	stomp.lg.Printf("connsess:%s common_connect_heart_beat_send hbsend:%d\n",
		stompConn.Session(),
		stompConn.SendTickerInterval())
	stomp.lg.Printf("connsess:%s common_connect_heart_beat_recv hbrecv:%d\n",
		stompConn.Session(),
		stompConn.ReceiveTickerInterval())

	stomp.stompConnection = stompConn

	return conn, err
}

func (stomp *StompClient) makeTcpConnection() (net.Conn, error) {

	n, err := net.Dial(stompngo.NetProtoTCP, stomp.Url.Host)

	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(strings.ToLower(stomp.Url.Scheme), "https") {
		//tc := new(tls.Config)
		tc := &tls.Config{}
		tc.InsecureSkipVerify = true
		tc.ServerName = stomp.getHost()

		sn := tls.Client(n, tc)
		err = sn.Handshake()
		if err != nil {
			return nil, err
		}
		return sn, nil
	}

	return n, nil
}

func (stomp *StompClient) getHost() string {
	host := stomp.Url.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return host
}
