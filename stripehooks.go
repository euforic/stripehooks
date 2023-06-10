package stripehooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Manager is the struct that holds the event handlers
type Manager struct {
	verify         bool // if true, verify the signature of the event
	endpointSecret string
	handlers       map[EventType]HandlerFunc
}

// HandlerFunc is the function signature for the event handler
type HandlerFunc func(*stripe.Event) error

// Option is the function signature for the Manager options
type Option func(*Manager)

// WithVerify sets the verify option
func WithVerify(endpointSecret string) Option {
	if endpointSecret == "" {
		panic("endpointSecret cannot be empty in WithVerify")
	}

	return func(m *Manager) {
		m.verify = true
		m.endpointSecret = endpointSecret
	}
}

// New returns a new Manager
func New(options ...Option) *Manager {
	m := Manager{
		handlers: make(map[EventType]HandlerFunc),
	}

	for _, option := range options {
		option(&m)
	}

	return &m
}

// Event registers a handler for the given event type
func (m *Manager) Handle(eventType EventType, fn HandlerFunc) {
	if m.handlers == nil {
		m.handlers = make(map[EventType]HandlerFunc)
		m.handlers[eventType] = fn
		return
	}

	m.handlers[eventType] = fn
}

// ProcessEvent processes the event
func (m *Manager) ProcessEvent(payload []byte, signatureHeader string) error {
	var event stripe.Event
	var err error

	if m.verify {
		event, err = webhook.ConstructEvent(payload, signatureHeader, m.endpointSecret)
		if err != nil {
			return fmt.Errorf("error verifying webhook signature: %v", err)
		}
	} else {
		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("error unmarshalling event: %v", err)
		}
	}

	fn, ok := m.handlers[EventType(event.Type)]
	if !ok {
		return nil
	}

	if err := fn(&event); err != nil {
		return fmt.Errorf("error handling event: %v", err)
	}

	return nil
}

// HttpHandler returns an http.HandlerFunc
func (m *Manager) HttpHandler(errFn func(err error)) http.HandlerFunc {
	if errFn != nil {
		// if errFn is nil, set it to a no-op function
		errFn = func(err error) {}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			httpErrorHandler(w, fmt.Errorf("error reading request body: %v", err))
			errFn(err)
			return
		}

		signatureHeader := r.Header.Get("Stripe-Signature")

		if err := m.ProcessEvent(payload, signatureHeader); err != nil {
			httpErrorHandler(w, fmt.Errorf("error processing event: %v", err))
			errFn(err)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			httpErrorHandler(w, fmt.Errorf("error writing response: %v", err))
			errFn(err)
			return
		}
	}
}

// httpErrorHandler handles the error and writes it to the response
func httpErrorHandler(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(err.Error())); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
