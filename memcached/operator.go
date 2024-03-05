package memcached

import (
	"os"
	"time"
)

type Operator interface {
	ProcessCommand([]byte) (<-chan []byte, <-chan error)
	// A key to identify the operator.
	Identifier() []byte
}

type Operation struct {
	Operator Operator
	Command  []byte
}

// Send command to operator asynchronously.
// Results are received on only one of the channels.
func (o Operation) Async() (<-chan []byte, <-chan error) {
	return o.Operator.ProcessCommand(o.Command)
}

// Call the [Operation.Async] and wait for timeout duration.
// If there is no return value within the duration, [os.ErrDeadlineExceeded] is returned.
// If timeout is zero, it will wait indefinitely.
func (o Operation) Timeout(timeout time.Duration) ([]byte, error) {
	responseChan, errorChan := o.Async()
	var timeoutChan <-chan time.Time
	if timeout != 0 {
		timeoutChan = time.After(timeout)
	}
	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errorChan:
		return nil, err
	case <-timeoutChan:
		return nil, os.ErrDeadlineExceeded
	}
}

// Send command to operator synchronously.
// This function is an alias for [Operation.Timeout] with zero timeout.
func (o Operation) Result() ([]byte, error) {
	return o.Timeout(0)
}

type operation struct {
	command      []byte
	responseChan chan<- []byte
	errorChan    chan<- error
}

func (o *operation) fail(err error) {
	o.errorChan <- err
	close(o.responseChan)
	close(o.errorChan)
}

func (o *operation) success(bytes []byte) {
	o.responseChan <- bytes
	close(o.responseChan)
	close(o.errorChan)
}
