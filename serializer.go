package websocket

import (
	"encoding/json"
	"google.golang.org/protobuf/proto"
)

type Serializer[T any] interface {
	Serialize(data T) ([]byte, error)
	Deserialize(data []byte, v *T) error
	Data() T
}

type JSONSerializer[T Message] struct {
	data T
}

func NewJSONSerializer[T Message](data T) *JSONSerializer[T] {
	return &JSONSerializer[T]{
		data: data,
	}
}

func (j *JSONSerializer[T]) Data() T {
	return j.data
}

func (j *JSONSerializer[T]) Serialize(data T) ([]byte, error) {
	return json.Marshal(data)
}

func (j *JSONSerializer[T]) Deserialize(data []byte, v *T) error {
	return json.Unmarshal(data, v)
}

type ProtocolSerializer[T proto.Message] struct {
	data T
}

func NewProtocolSerializer[T proto.Message](data T) *ProtocolSerializer[T] {
	return &ProtocolSerializer[T]{
		data: data,
	}
}

func (p *ProtocolSerializer[T]) Data() T {
	return p.data
}

func (p *ProtocolSerializer[T]) Serialize(data T) ([]byte, error) {
	return proto.Marshal(data)
}

func (p *ProtocolSerializer[T]) Deserialize(data []byte, v *T) error {
	var obj T
	if err := proto.Unmarshal(data, obj); err != nil {
		return err
	}
	v = &obj
	return nil
}
