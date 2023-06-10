# Stripehooks

[![GoDoc](https://pkg.go.dev/badge/github.com/euforic/stripehooks)](https://pkg.go.dev/github.com/euforic/stripehooks)

The `stripehooks` package provides a set of tools for processing Stripe webhook events in a Go application.

## Types and Functions

### `Manager`

The `Manager` struct holds the handlers for processing Stripe webhook events. It provides several methods for processing events and handling HTTP requests.

### `HandlerFunc`

The `HandlerFunc` type is a function that takes a Stripe event and returns an error. It is used to process Stripe events.

### `Option`

The `Option` type is a function that modifies a `Manager` instance. It is used to set options on a `Manager`.

## Methods

### `New(options ...Option) *Manager`

The `New` function returns a new `Manager` instance with the provided options.

### `Handle(eventType EventType, fn HandlerFunc)`

The `Handle` method registers a handler function for the given event type.

### `ProcessEvent(payload []byte, signatureHeader string) error`

The `ProcessEvent` method processes a Stripe event. It verifies the event signature if the `verify` option is set, unmarshals the event, and calls the registered handler for the event type.

### `HttpHandler(errFn func(err error)) http.HandlerFunc`

The `HttpHandler` method returns an HTTP handler function that reads the request body, processes the event, and writes the response.
An optional errFn can be passed in to allow for logging errors that may be encounter while processing the incomming webhook.

## Options

### `WithVerify(endpointSecret string) Option`

The `WithVerify` function returns an option that sets the `verify` and `endpointSecret` fields on a `Manager`.

### `WithHttpErrorHandler() Option`

The `WithHttpErrorHandler` function returns an option that allows the developer to process errors that occure in the `HttpHandler`

## Example Usage

```go
manager := stripehooks.New(
	stripehooks.WithVerify("your_endpoint_secret"),
)

manager.Handle(stripehooks.ChargeSucceeded, func(event *stripe.Event) error {
	// handle event
	return nil
})

http.HandleFunc("/webhook", manager.HttpHandler())
http.ListenAndServe(":8080", nil)
```

In this example, a new `Manager` is created with the `WithVerify` option, a handler for "charge.succeeded" events is registered, and the manager's HTTP handler function is registered with the net/http package to handle requests at the "/webhook" path.
