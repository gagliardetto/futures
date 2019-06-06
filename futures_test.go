package futures

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	f := New()

	assert.NotNil(t, f.(*simpleFutures).futures, "New().futures must NOT be nil")
	assert.NotNil(t, f.(*simpleFutures).mu, "New().mu must NOT be nil")
}

func Test_Ask(t *testing.T) {
	f := New()
	key := "one"
	go func() {
		_, _ = f.Ask(key)
	}()

	time.Sleep(time.Second)

	v, ok := f.(*simpleFutures).futures[key]
	assert.True(t, ok, "an array of listening channels must have been created for the key")
	assert.Len(t, v, 1, "f.futures[key] array must contain one listener channel")
}

func Test_Answer(t *testing.T) {
	f := New()
	key := "one"
	answerValue := 33
	go func() {
		time.Sleep(time.Second * 2)
		howManyReceived := f.Answer(key, answerValue, nil)
		assert.Equal(t, 1, howManyReceived, "one listener should have received the answer")
	}()

	a, err := f.Ask(key)
	assert.EqualValues(t, answerValue, a, "what is answered and received should match")
	assert.Nil(t, err, "err should be nil")
}
