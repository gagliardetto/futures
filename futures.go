package futures

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

// ErrTimeout is the error returned in case of an AskWithTimeout
var ErrTimeout = errors.New("timeout exceeded")

type answer struct {
	err error
	val interface{}
}

// simpleFutures is the struct used to receive and ask for values
type simpleFutures struct {
	futures map[interface{}][]chan answer
	mu      *sync.RWMutex
}

// New returns a new simpleFutures that can be used to send and receive values
func New() Futures {
	return &simpleFutures{
		futures: make(map[interface{}][]chan answer),
		mu:      &sync.RWMutex{},
	}
}

// Futures is the interface used to receive and ask for values
type Futures interface {
	Answer(key interface{}, val interface{}, err error) int
	Ask(key interface{}) (interface{}, error)
	AskWithTimeout(key interface{}, timeout time.Duration) (interface{}, error)

	AskWithTimeoutAndPostSubscriptionCallback(
		key interface{},
		timeout time.Duration,
		postSubCallback func(),
	) (interface{}, error)
}

// Answer sends an answer to any listener currently listening to the key;
// when answering, the list of listeners is locked, and after all listeners
// have received the answers, the list of listeners is deleted for that key.
func (f *simpleFutures) Answer(key interface{}, val interface{}, err error) int {
	message := answer{
		val: val,
		err: err,
	}

	v, ok := f.futures[key]
	if !ok {
		return 0
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	written := 0
	for i := range v {
		listener := v[i]
		select {
		case listener <- message:
			written++
			close(listener)
		default:
			continue
		}
	}

	f.futures[key] = nil

	return written
}

// Ask waits for an answer for the provided key (indefinitely)
func (f *simpleFutures) Ask(key interface{}) (interface{}, error) {

	c := f.subscribe(key)

	t := <-c

	return t.val, t.err
}

// AskWithTimeout waits for an answer for the provided key until the timeout passed (which returns ErrTimeout)
func (f *simpleFutures) AskWithTimeout(key interface{}, timeout time.Duration) (interface{}, error) {
	select {
	case t := <-f.subscribe(key):
		return t.val, t.err

	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

// AskWithTimeout waits for an answer for the provided key until the timeout passed (which returns ErrTimeout);
// the callback is called right after the subscription is done; the callback MUST BE NON-BLOCKING.
func (f *simpleFutures) AskWithTimeoutAndPostSubscriptionCallback(
	key interface{},
	timeout time.Duration,
	postSubCallback func(),
) (interface{}, error) {

	sub := f.subscribe(key)

	// execute the callback:
	postSubCallback()

	// wait for answer, or timeout:
	select {
	case t := <-sub:
		return t.val, t.err

	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (f *simpleFutures) subscribe(key interface{}) chan answer {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}

	newListener := make(chan answer)

	f.addListener(key, newListener)

	return newListener
}

func (f *simpleFutures) addListener(key interface{}, newListener chan answer) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.futures[key] == nil {
		f.futures[key] = make([]chan answer, 0)
	}

	f.futures[key] = append(f.futures[key], newListener)
}
