package websocket

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
)

type Serializer[T any] interface {
	Serialize(data T) ([]byte, error)
	Deserialize(data []byte, v *T) error
}

type JSONSerializer[T any] struct{}

func NewJSONSerializer[T any]() *JSONSerializer[T] {
	return &JSONSerializer[T]{}
}

func (j *JSONSerializer[T]) Serialize(data T) ([]byte, error) {
	return json.Marshal(data)
}

func (j *JSONSerializer[T]) Deserialize(data []byte, v *T) error {
	return json.Unmarshal(data, v)
}

type ProtocolSerializer[T any] struct{}

func NewProtocolSerializer[T any]() *ProtocolSerializer[T] {
	return &ProtocolSerializer[T]{}
}

func (p *ProtocolSerializer[T]) Serialize(data T) ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *ProtocolSerializer[T]) Deserialize(data []byte, v *T) error {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, v)
	if err != nil {
		return errors.New("failed to deserialize protocol data")
	}
	return nil
}
