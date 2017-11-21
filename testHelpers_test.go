package eventuate_test

import (
	"net/http"
	"testing"

	//"github.com/shopcookeat/eventuate-client-golang"
	"github.com/stretchr/testify/assert"
)

//type RewriteTransport struct {
//	Transport http.RoundTripper
//}
//
//func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, Error) {
//	req.URL.Scheme = "http"
//	if t.Transport == nil {
//		return http.DefaultTransport.RoundTrip(req)
//	}
//	return t.Transport.RoundTrip(req)
//}

func assertMethod(t *testing.T, expectedMethod string, req *http.Request) {
	assert.Equal(t, expectedMethod, req.Method)
}

func assertNoError(t *testing.T, err error) {
	assert.Equal(t, nil, err)
}

//
//func readNEventsSequentially(sub *eventuate.Subscription, count int) []*eventuate.StompEvent {
//
//	result := make([]*eventuate.StompEvent, count)
//	for idx := range result {
//		evt, err := sub.ReadEvent()
//		if err != nil {
//
//		}
//		result[idx] = evt
//	}
//
//	return result
//}
