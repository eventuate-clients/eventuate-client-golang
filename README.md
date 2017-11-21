# Golang client for Eventuate API

## Install

To install this package run
```bash
go get https://github.com/eventuate-clients/eventuate-client-golang
```
or clone/create this package as git submodule for your projects vendor directory.

## Usage

To use the library, import it with the statement:
```go
import  eventuate "https://github.com/eventuate-clients/eventuate-client-golang"
```


## Definitions

### Aggregates

In the Eventuate programming model, the majority of your application's business logic is implemented by aggregates. An aggregate does two things:

* Processes commands and returns events, which leaves the state of the aggregate unchanged.
* Consumes events, which updates its state.

The [Eventuate client framework for Golang](https://https://github.com/eventuate-clients/eventuate-client-golang) relies on a simple way to define aggregates.

It has to be a public type, like so:
```go
type FooBarAggregate struct {
	state string
}
```

accompanied by an instantiation function:
```go
func NewFooBarAggregate() *FooBarAggregate {
	return &FooBarAggregate{
		state: "New"}
}
```

and its own business logic:

```go
func (todo *FooBarAggregate) Foo(foo string) {
	todo.state = fmt.Sprintf("%v;%v:%v", todo.state, "foo", foo)
}
func (todo *FooBarAggregate) Bar(bar string) {
	todo.state = fmt.Sprintf("%v;%v:%v", todo.state, "bar", bar)
}
```

to top it all, you have to prepare a fully-qualified name (FQN) for this aggregate:

```go
const FOOBAR_ENTITY = "golang.FooBarApplication.examples.FooBarEntity"
```

### Commands and events

As mentioned above, the Aggregate processes commands (returning events and leaving its state unchanged) and consumes events (which update its state).

A command is a desired action to be performed on the instance of an up-to-date entity. Once processed, it generates one or several, or none, events ('increased this', 'decreased that', 'signaled an error', etc.). A command is not supposed to get serialized or passed outside of the boundaries of a local code.

Events, on the other hand, are designed to be serialized and passed on to the Eventuate server. They carry bits of data, which, when applied in order, recreate an entity's instance in its current state.


#### Commands

Commands are defined as simple `struct`s whose names end with `Command`:
```go
type FooCommand struct {
	Foo string
}

type BarCommand struct {
	Bar string
}
```

Your aggregate needs to be able to process them. Either with specific methods:
```go
func (foobar *FooBarAggregate) ProcessFooCommand(cmd *FooCommand) []eventuate.Event {
    // logic based of the foobar's current state
	return []eventuate.Event{
		newFooEvent(cmd.Foo)}
}
func (foobar *FooBarAggregate) ProcessBarCommand(cmd *BarCommand) []eventuate.Event {
    // logic based of the foobar's current state
	return []eventuate.Event{
		newBarEvent(cmd.Bar)}
}

func newFooEvent(foo string) *FooEvent {
	return &FooEvent{
		Foo: foo}
}

func newBarEvent(bar string) *BarEvent {
	return &BarEvent{
		Bar: bar}
}
```
.. or with a general one:
```go
func (foobar *FooBarAggregate) ProcessCommand(command eventuate.Command) []eventuate.Event {
	switch cmd := command.(type) {

	case *FooCommand:
		{
			return []eventuate.Event{
				newFooEvent(cmd.Foo)}
		}
	case *BarCommand:
		{
			return []eventuate.Event{
				newBarEvent(cmd.Bar)}
		}
	}

	return []eventuate.Event{}
}

```

*Important!* Note, how specific command-processing methods are named `Process`XYZ`Command` and return slices of `eventuate.Event`. Also note how general command-processing method is named `ProcessCommand` and accepts a type which is an alias to `interface{}`, `eventuate.Command`. If specific methods are present, they take precedence over the general one. Thus, there is no need to have the latter if all specific methods are defined.

#### Events

Events are defined as simple `struct`s whose names end with `Event` and whose fields (as a rule) have JSON tags as (de)serialization guidelines:
```go
type FooEvent struct {
	Foo string `json:"foo"`
}

type BarEvent struct {
	Bar string `json:"bar"`
}
```

alongside with the defined Event types there must be prepared their fully-qualified names (FQN):

```go
const FOOBAR_FOO_EVENT = "golang.FooBarApplication.examples.FooBarEntity.Events.FooEvent"
const FOOBAR_BAR_EVENT = "golang.FooBarApplication.examples.FooBarEntity.Events.BarEvent"
```

Your aggregate needs to be able to apply these events with either specific methods:
```go
func (foobar *FooBarAggregate) ApplyFooEvent(evt *FooEvent) *FooBarAggregate {
	foobar.Foo(evt.Foo)
	return todo
}
func (foobar *FooBarAggregate) ApplyBarEvent(evt *BarEvent) *FooBarAggregate {
	foobar.Bar(evt.Bar)
	return todo
}
```
or with a general one:
```go
func (foobar *FooBarAggregate) ApplyEvent(evt eventuate.Event) *FooBarAggregate {
	switch t := evt.(type) {
	case *FooEvent:
		{
			foobar.Foo(t.Foo)
		}
	case *BarEvent:
		{
			foobar.Bar(t.Bar)
		}
	}
	return todo
}
```

*Important!* Note, how specific event-application methods are named `Apply`XYZ`Event` and return a reference to the same receiver (`*FooBarAggregate` here). Also note how general event-application method is named `ApplyEvent` and accepts a type which is an alias to `interface{}`, `eventuate.Event`. If specific methods are present, they take precedence over the general one. Thus, there is no need to have the latter if all specific methods are defined.

### Configuring metadata for an aggregate

The rules outlined above (marked with the '*Important!*') are enforced and used during command-processing and event-application with the help of an instance of `eventuate.AggregateMetadata`. This metadata-holding object is crucial for all operations with the Eventuate store, so it is important to have it handy prior to establishing connection to Eventuate:

```go
aggregateMetadata, metaErr := eventuate.CreateAggregateMetadata(NewFooBarAggregate, FOOBAR_ENTITY)
// check for and handle errors
```
Note that we pass the aggregate's instantiation function and its fully-qualified name. Refer back to the section [Aggregates](#aggregates).

#### Event (de)serialization hints

Since events need to be passed over the network and be correctly serialized and de-serialized along with their fully-qualified names, we need to provide FQN-Type correspondence. Thus:
```go
aggregateMetadata.RegisterEventType(FOOBAR_FOO_EVENT, &FooEvent{})
aggregateMetadata.RegisterEventType(FOOBAR_BAR_EVENT, &BarEvent{})
```
Failure to do so will prevent a correct deserialization of events and, as a result, their application. We pass a sample of the instantiated event type following its FQN.

## Working with the Eventuate server's API

### Creating a REST client

To create a REST Client, prepare a variable of `eventuate.AggregateCrud` type:
```go
var client eventuate.AggregateCrud
```
and use Eventuate client builder API:
```go
client, _ = eventuate.ClientBuilder().BuildREST()
// check for and handle errors
```

This defaults to connecting to the Eventuate REST API server's address `http://api.eventuate.io` (and STOMP API server's address `https://dev.eventuate.io:61615`), picking credentials from the environment variables `EVENTUATE_API_KEY_ID` and `EVENTUATE_API_KEY_SECRET`, setting a namespace for operations to `"default"`, and silent log mode. If these defaults need to be changed use the following chainable builder option methods:

#### `WithUrl(serverUrl)`

To change default Eventuate REST API server's address. (Default is `http://api.eventuate.io`)
```go
client, _ = eventuate.ClientBuilder().WithUrl("https://dev.eventuate.io").BuildREST()
// check for and handle errors
```

#### `WithSpace(space)`

To change default operations namespace. (Default is `"default"`)
```go
const FOOBAR_NS = "FooBarNamespace"
client, _ = eventuate.ClientBuilder().WithSpace(FOOBAR_NS).BuildREST()
// check for and handle errors
```

#### `WithCredentials(apiKeyId, apiKeySecret)`

To provide credentials directly instead of relying on reading the environment variables `EVENTUATE_API_KEY_ID` and `EVENTUATE_API_KEY_SECRET`
```go
client, _ = eventuate.ClientBuilder().WithCredentials("ABCD0987..", "FEDCBA010...01234").BuildREST()
// check for and handle errors
```


#### `WithDevMode(devMode)`

```go
client, _ = eventuate.ClientBuilder().WithDevMode(eventuate.Verbose).BuildREST()
// check for and handle errors
```


### Create an AggregateRepository

Once the REST API Client instance is ready, an Aggregate repository instantiation is due:
```go
// var repo *AggregateRepository
repo := eventuate.NewAggregateRepository(&client)
```
The creation of an entity, its retrieving and updating is done against the repository instance (`repo` here):

### Saving aggregate

```go
entity, _ := repo.Save(aggregateMetadata, &FooCommand{
    Foo: "FooString"})
// check for and handle errors

entityId := entity.EntityId // id is used to reference a newly created entity
```

### Updating aggregate

```go
entity, _ = repo.Update(aggregateMetadata, entityId, &BarCommand{
    Bar: "BarString"})
// check for and handle errors

```
### Finding aggregate

```go
locatedEntity, _ := repo.Find(aggregateMetadata, entityId)
// check for and handle errors

entityInstance := locatedEntity.EntityInstance
```

### STOMP

#### Creating a STOMP client

Establishing a STOMP connection uses the same `eventuate.ClientBuilder()` API with the only difference of calling `BuildSTOMP()` in the end.

##### `BuildSTOMP()` (for STOMP)

To instantiate Eventuate STOMP API Client instance:
```go
var stomp *StompClient
stomp, _ = eventuate.ClientBuilder().BuildSTOMP()
// check for and handle errors
```

##### `WithStompUrl(serverUrl)` (for STOMP)

To change default Eventuate STOMP API server's address. (Default is `https://api.eventuate.io:61614`)
```go
stomp, _ := eventuate.ClientBuilder().WithStompUrl("https://dev.eventuate.io:61615").BuildSTOMP()
// check for and handle errors
```

#### Subscription manager

Subscription manager instantiation:
```go
var (
        sm *eventuate.SubscriptionManager
)
sm, _ := eventuate.NewSubscriptionManager(stomp)
// check for and handle errors
```

#### Subscription

Subscribing for events:
```go
subscriberId := "..." // generate UID or use your constant
entityEventTypeMap := eventuate.EventResultHandlerMap{
    FOOBAR_ENTITY: {
        FOOBAR_FOO_EVENT: testResult.generalResultEventHandler,
        FOOBAR_BAR_EVENT: testResult.generalResultEventHandler}}

sub, _ := sm.Subscribe(subscriberId, entityEventTypeMap, false)
// check for errors first
```

#### Event FQN-Type registration

Event types registration:
```go
sm.RegisterEventType(FOOBAR_FOO_EVENT, FooEvent{})
sm.RegisterEventType(FOOBAR_BAR_EVENT, BarEvent{})
```

This is obviously best done prior to `.Subscribe(..)`.

#### Event handler

Event handler must implement the type `eventuate.EventResultHandler`, which looks like:
```go
type EventResultHandler func(*DeserializedEvent) *Settler
```

The `eventuate.DeserializedEvent` contains these useful fields: `Id`, `EntityId`, `EntityType`, `EventType`, and `EventData`.

The `eventuate.Settler` interface is a Javascripts's rough equivalent of a `Promise`, and Java's equivalent of a `CompletableFuture<T>`:
```go
type Settler interface {
	IsSettled() bool
	Settle(interface{}, error)
	GetValue() (interface{}, error)
}
```

You can use the implementation supplied by the Library: `eventuate.FutureResult` and the instantiation helper functions: `eventuate.NewPassedFutureResult(val interface{})` and `eventuate.NewFailedFutureResult(err error)`. (**Important!** Please be careful with calling `.GetValue()` without checking for settled-ness (`.IsSettled()`) on instances since this may easily block your code in async scenarios.)

Event handler sample (referred to from the subscription snippet above):
```go
type result struct {
	count int
}

func (rslt *result) generalResultEventHandler(evt *eventuate.DeserializedEvent) *eventuate.Settler {
	log.Printf("Event handler for: %#v\n", evt)
	switch evtData := evt.EventData.(type) {
	case *FooEvent:
		{
			rslt.count += 1
			log.Printf("Data is of type `*FooEvent`, value: %v\n", evtData.Foo)
		}
	case *BarEvent:
		{
			rslt.count += 8
			log.Printf("Data is of type `*BarEvent`, value: %v\n", evtData.Bar)
		}
	}
	return eventuate.NewPassedFutureResult(true)
}
```

Expect async work before exiting.


## Run tests

To check tests you need to run ```go test``` in the project root. Before you run the tests export your API token id and API token secret, for example:

```
export EVENTUATE_API_KEY_ID=key_id
export EVENTUATE_API_KEY_SECRET=key_secret
```
