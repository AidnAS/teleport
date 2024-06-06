package encoder

import (
	"fmt"
)

// Encoder is the interface that wraps the basic Encode method.
// Anything implementing Encoder can be turned into bytes using Kafka's encoding rules.
type Encoder interface {
	Encode(pe packetEncoder) error
}

type encoderWithHeader interface {
	Encoder
	headerVersion() int16
}

// PacketEncodingError is returned from a failure while encoding a Kafka packet. This can happen, for example,
// if you try to encode a string over 2^15 characters in length, since Kafka's encoding rules do not permit that.
type PacketEncodingError struct {
	Info string
}

func (err PacketEncodingError) Error() string {
	return fmt.Sprintf("kafka: error encoding packet: %s", err.Info)
}

const MaxRequestSize int32 = 100 * 1024 * 1024

// Encode takes an Encoder and turns it into bytes while potentially recording metrics.
func Encode(e Encoder) ([]byte, error) {
	if e == nil {
		return nil, nil
	}

	var prepEnc prepEncoder
	var realEnc realEncoder

	err := e.Encode(&prepEnc)
	if err != nil {
		return nil, err
	}

	if prepEnc.length < 0 || prepEnc.length > int(MaxRequestSize) {
		return nil, PacketEncodingError{fmt.Sprintf("invalid request size (%d)", prepEnc.length)}
	}

	realEnc.raw = make([]byte, prepEnc.length)
	// realEnc.registry = metricRegistry
	err = e.Encode(&realEnc)
	if err != nil {
		return nil, err
	}

	return realEnc.raw, nil
}
