package memcached

type Operator interface {
	AsyncOperation([]byte) (<-chan []byte, <-chan error)
	Operation([]byte) ([]byte, error)
}

type operation struct {
	asciiCommand []byte
	responseChan chan<- []byte
	errorChan    chan<- error
}

func (self *operation) errorResponse(err error) {
	self.errorChan <- err
	close(self.responseChan)
	close(self.errorChan)
}

func (self *operation) bytesResponse(bytes []byte) {
	self.responseChan <- bytes
	close(self.responseChan)
	close(self.errorChan)
}
