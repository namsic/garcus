package memcached

type Operator interface {
	AsyncOperation([]byte) (<-chan []byte, <-chan error)
	Operation([]byte) ([]byte, error)
}

type operation struct {
	command      []byte
	responseChan chan<- []byte
	errorChan    chan<- error
}

func (o *operation) errorResponse(err error) {
	o.errorChan <- err
	close(o.responseChan)
	close(o.errorChan)
}

func (o *operation) bytesResponse(bytes []byte) {
	o.responseChan <- bytes
	close(o.responseChan)
	close(o.errorChan)
}
