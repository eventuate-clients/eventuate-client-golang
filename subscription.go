package eventuate

import (
	"encoding/json"
	"fmt"
	"github.com/gmallard/stompngo"
	loglib "github.com/shopcookeat/eventuate-client-golang/logger"
	"strings"
	"sync"
	"time"
)

// Subscription is the struct for subscription
type Subscription struct {
	sync.RWMutex
	client                   *StompClient
	Id                       string // Unique Subscription EntityId
	isActive                 bool
	receiptChannel           <-chan stompngo.MessageData
	incomingEvent            chan StompEvent // Holds the parsed STOMP message data
	subscriptionErrors       chan error
	ackEvent                 chan *StompEvent // Allows to ack() an Event
	unsubscribeFn            func() error
	reqStop                  chan bool
	reqCleanup               chan bool
	pendingsCountReqChannel  chan bool
	pendingsCountRespChannel chan int
	eventHandler             *EventResultHandler
	ll                       loglib.LogLevelEnum
	lg                       loglib.Logger
	lmu                      sync.Mutex
}

type pendingAcknowledge struct {
	EventID   Int128
	Acked     bool
	AckHeader string
	DebugInfo interface{}
}

type StompEventWithError struct {
	Event StompEvent
	Error error
}

func newSubscription(
	uid string,
	conn *stompngo.Connection, //Acker,
	receiptChannel <-chan stompngo.MessageData,
	eventHandler *EventResultHandler) *Subscription {

	sub := &Subscription{
		//stomp:                   stomp,
		Id:                       uid,
		isActive:                 true,
		receiptChannel:           receiptChannel,
		incomingEvent:            make(chan StompEvent),
		subscriptionErrors:       make(chan error, 16),
		ackEvent:                 make(chan *StompEvent, 16),
		reqStop:                  make(chan bool),
		reqCleanup:               make(chan bool),
		pendingsCountReqChannel:  make(chan bool),
		pendingsCountRespChannel: make(chan int),
		eventHandler:             eventHandler,
		ll:                       loglib.Silent,
		lg:                       loglib.NewLogger(loglib.Silent)}

	go func(sub *Subscription) {
		for {
			if conn.Connected() {
				sub.lg.Printf("STOMP Connection alive (Sub. #%v)", sub.Id)
				time.Sleep(time.Duration(1) * time.Minute)
				continue
			}
			sub.lg.Printf("STOMP Connection SEVERED. (Sub. #%v)", sub.Id)
		}

	}(sub)

	var pchan chan pendingAcknowledge = make(chan pendingAcknowledge)

	go func(sub *Subscription, pchan chan pendingAcknowledge) {
		for md := range sub.receiptChannel {
			var (
				stompEvent StompEvent
				err        error
			)

			if md.Error != nil {
				sub.subscriptionErrors <- err
				sub.reqStop <- true

				continue
			}

			if md.Message.Command != stompngo.MESSAGE {
				sub.lg.Printf("Bad frame: %v", md.Message.Command)
				sub.subscriptionErrors <- AppError("Bad frame: %v", md.Message.Command)
				continue
			}

			err = json.Unmarshal(md.Message.Body, &stompEvent)
			if err != nil {
				sub.lg.Printf("Cannot unmarshal event body: %v with error: %v", string(md.Message.Body), err.Error())
				sub.subscriptionErrors <- err
				continue
			}

			ackHeaderId := md.Message.Headers.Value("ack")
			pchan <- pendingAcknowledge{
				EventID:   stompEvent.Id,
				Acked:     false,
				AckHeader: ackHeaderId,
				DebugInfo: stompEvent.String()}

			sub.incomingEvent <- stompEvent
		}

	}(sub, pchan)

	go func(acker Acker, sub *Subscription, pchan chan pendingAcknowledge) {
		sub.lg.Println("Subscription go-routine started")
		defer sub.lg.Println("Subscription go-routine finished")
		defer sub.Unsubscribe()

		pendings := make([]pendingAcknowledge, 0)

		for {

			if !sub.IsActive() {
				return
			}

			select {
			case <-sub.reqStop:
				{
					sub.lg.Println("reqStop chan, will terminate")
					return
				}

			case pack := <-pchan:
				{
					pendings = append(pendings, pack)
				}
			case event := <-sub.ackEvent:
				{
					sub.lg.Printf("newSubscription.g3: Received event via ackEvent channel: %v\n", event)

					needAcks, nextPendings := sub.getAcks(pendings, event)
					sub.lg.Printf("newSubscription.g3: Pendings count: %v, Need Acks count: %v, Next pendings count: %v\n", len(pendings), len(needAcks), len(nextPendings))

					for _, pending := range needAcks {

						ackHeaders := stompngo.Headers{
							"id", pending.AckHeader}

						sub.lg.Printf("newSubscription.g3: Before calling stompngo.Ack: %v\n", ackHeaders)
						err := acker.Ack(ackHeaders)
						if err != nil {
							sub.lg.Printf("newSubscription.g3: Error in stompngo.Ack(ackHeaders): %s\n%v",
								pending.AckHeader, err)

							sub.subscriptionErrors <- AppError(
								"Error in StompConnection.Ack(ackHeaders): %s\n%v",
								pending.AckHeader, err)
						} else {
							sub.lg.Printf("newSubscription.g3: After calling stompngo.Ack (OK): %v\n", ackHeaders)
						}
					}

					pendings = nextPendings
				}

			case <-sub.reqCleanup:
				{
					sub.lg.Println("reqCleanup channel (cleaning & closing)")

					// cleanup
					close(sub.reqStop)
					close(sub.reqCleanup)
					close(sub.incomingEvent)
					close(sub.subscriptionErrors)
					close(sub.ackEvent)
					close(sub.pendingsCountReqChannel)
					close(sub.pendingsCountRespChannel)
					sub.eventHandler = nil
					return
				}
			case err := <-sub.subscriptionErrors:
				{
					if err != nil && sub.lg != nil {
						sub.lg.Printf("subscriptionErrors channel (proceeding): %v", err.Error())
					}
				}

			case <-sub.pendingsCountReqChannel:
				{
					sub.pendingsCountRespChannel <- len(pendings)
				}
			}
		}

	}(conn, sub, pchan)

	return sub
}

func (sub *Subscription) SetLogLevel(logLevel loglib.LogLevelEnum) {
	sub.lmu.Lock()
	sub.ll = logLevel
	sub.lg = loglib.NewLogger(logLevel)
	sub.lmu.Unlock()
}

func (sub *Subscription) IsActive() bool {
	sub.RWMutex.RLock()
	defer sub.RWMutex.RUnlock()
	return sub.isActive
}

func (sub *Subscription) Unsubscribe() error {
	sub.lg.Printf("Subscription.Unsubscribe()")

	sub.RWMutex.Lock()
	defer sub.RWMutex.Unlock()
	if sub.isActive {
		err := sub.unsubscribeFn()
		sub.isActive = false
		if err == nil {
			sub.reqCleanup <- true
		}
		return err
	}
	return nil
}

// ReadEvent is the function to read Event from subscription
func (sub *Subscription) ReadEvent() (*StompEvent, error) {
	evt, ok := <-sub.incomingEvent
	if ok {
		return &evt, nil
	}
	return nil, AppError("Cannot read from a closed Subscription")
}

// ReadEventNonblocking is a function that doesn't block reading of events
func (sub *Subscription) ReadEventNonblocking() (*StompEvent, error, bool) {
	select {
	case err, errOk := <-sub.subscriptionErrors:
		{
			if errOk {
				return nil, err, true
			}
			return nil, AppError("Cannot read from a closed Subscription"), true
		}
	case evt, ok := <-sub.incomingEvent:
		{
			if ok {
				return &evt, nil, true
			}
			return nil, AppError("Cannot read from a closed Subscription"), true
		}
	default:
		{
			return nil, nil, false
		}
	}
}

// AcknowledgeEvent is the function to ack an Event
func (sub *Subscription) AcknowledgeEvent(event *StompEvent) {
	sub.ackEvent <- event
	return
}

// FetchPendingsCount is an internal function to check the count of events awaiting their acks
func (sub *Subscription) FetchPendingsCount() int {
	sub.pendingsCountReqChannel <- true
	return <-sub.pendingsCountRespChannel
}

func (sub *Subscription) getAcks(pendings []pendingAcknowledge, event *StompEvent) (needAcks []pendingAcknowledge, nextPendings []pendingAcknowledge) {

	sub.lg.Printf("getAcks. Debug info: \n%v", sub.getDebugInfo(pendings, event.Id))
	sub.lg.Println(sub.listUnacked(pendings))
	i := sub.findEventIndex(pendings, event.Id)

	if i == -1 {
		sub.lg.Printf("getAcks. Event with Id %v missing from pendingAcknowledges", event.Id)
		return make([]pendingAcknowledge, 0), pendings
	}

	sub.lg.Printf("getAcks. Event with Id %v present in pendingAcknowledges at #%v", event.Id, i)

	// we mark the event first
	pendings[i].Acked = true

	i = findFirstNotAckedIndex(pendings)

	needAcks = pendings[0:i]
	sub.lg.Printf("getAcks. Count of events that need acks: %v", len(needAcks))

	nextPendings = pendings[i:]
	sub.lg.Printf("getAcks. Debug info (nextPendings): \n%v", sub.getDebugInfo(nextPendings, event.Id))

	reallocatedNext := make([]pendingAcknowledge, len(nextPendings))
	copy(reallocatedNext, nextPendings)
	nextPendings = reallocatedNext

	return
}

func (sub *Subscription) findEventIndex(pendings []pendingAcknowledge, eventID Int128) int {
	for idx, p := range pendings {
		if p.EventID == eventID {
			return idx
		}
	}
	return -1
}

func (sub *Subscription) getDebugInfo(pendings []pendingAcknowledge, eventID Int128) string {
	result := make([]string, len(pendings))
	for idx, p := range pendings {
		result[idx] = "-"
		if p.Acked {
			result[idx] = "+"
		} else if p.EventID == eventID {
			result[idx] = "*"
		}
		if (idx != 0) && (idx%10 == 0) {
			result[idx] = fmt.Sprintf(".%v", result[idx])
		}
	}
	return strings.Join(result, "")
}

func (sub *Subscription) listUnacked(pendings []pendingAcknowledge) string {
	result := make([]string, 0)
	for _, p := range pendings {
		if !p.Acked {
			result = append(result, fmt.Sprintf("%v", p.DebugInfo))
		}
	}
	return strings.Join(result, "\n")
}

func findFirstNotAckedIndex(pendings []pendingAcknowledge) int {
	for idx, p := range pendings {
		if !p.Acked {
			return idx
		}
	}
	return len(pendings)
}
