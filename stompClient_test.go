package eventuate_test

import (
	"log"
	//"net/url"
	_ "testing"
	"github.com/eventuate-clients/eventuate-client-golang"
)

func ExampleNewStompClient() {
	client, clientErr := eventuate.ClientBuilder().BuildSTOMP()
	if clientErr != nil {
		log.Fatal(clientErr)
	}

	log.Println(client.Url.String())
}
