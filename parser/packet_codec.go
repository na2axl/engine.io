package parser

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type packetCodec interface {
	decode(data []byte) (*Packet, error)
	encode(packet *Packet) ([]byte, error)
	writeTo(writer io.Writer, packet *Packet) error
}

var (
	stringEncoder packetCodec = new(strCodec)
	binaryEncoder packetCodec = new(binCodec)
	base64Encoder packetCodec = new(b64Codec)
)

type binCodec struct {
}

func (p *binCodec) decode(data []byte) (*Packet, error) {
	if data == nil || len(data) < 1 {
		return nil, errors.New("packet bytes is empty")
	}
	t := PacketType(data[0])
	switch t {
	default:
		return nil, fmt.Errorf("invalid packet type: %d", t)
	case OPEN, CLOSE, PING, PONG, MESSAGE, UPGRADE, NOOP:
		return NewPacketCustom(t, data[1:], BINARY), nil
	}
}

func (p *binCodec) writeTo(writer io.Writer, packet *Packet) error {
	if _, err := writer.Write([]byte{byte(packet.Type)}); err != nil {
		return err
	}
	if packet.Data == nil || len(packet.Data) < 1 {
		return nil
	}
	_, err := writer.Write(packet.Data)
	return err
}

func (p *binCodec) encode(packet *Packet) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := p.writeTo(bf, packet); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

type strCodec struct {
}

func (p *strCodec) decode(data []byte) (*Packet, error) {
	if data == nil || len(data) < 1 {
		return nil, errors.New("packet bytes is empty")
	}
	t, err := convertCharToType(data[0])
	if err != nil {
		return nil, err
	}
	return NewPacketCustom(t, data[1:], 0), nil
}

func (p *strCodec) writeTo(writer io.Writer, packet *Packet) error {
	var t byte
	var err error
	if t, err = convertTypeToChar(packet.Type); err != nil {
		return err
	}
	if _, err = writer.Write([]byte{t}); err != nil {
		return err
	}
	_, err = writer.Write(packet.Data)
	return err
}

func (p *strCodec) encode(packet *Packet) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := p.writeTo(bf, packet); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

type b64Codec struct {
}

func (p *b64Codec) decode(data []byte) (*Packet, error) {
	l := len(data)
	if l < 1 {
		return nil, errors.New("packet bytes is empty")
	}
	if l < 2 {
		t, err := convertCharToType(data[0])
		if err != nil {
			return nil, err
		}
		return NewPacketCustom(t, make([]byte, 0), BINARY), nil
	}
	if data[0] != 'b' {
		return nil, fmt.Errorf("invalid b64 packet: %s", data)
	}
	if t, err := convertCharToType(data[1]); err != nil {
		return nil, err
	} else if data, err := base64.StdEncoding.DecodeString(string(data[2:])); err != nil {
		return nil, err
	} else {
		return NewPacketCustom(t, data, BINARY), nil
	}
}

func (p *b64Codec) writeTo(writer io.Writer, packet *Packet) error {
	var t byte
	var err error
	if t, err = convertTypeToChar(packet.Type); err != nil {
		return err
	}
	if packet.Data == nil || len(packet.Data) < 1 {
		if _, err = writer.Write([]byte{t}); err != nil {
			return err
		}
		return nil
	}
	if _, err = writer.Write([]byte{'b'}); err != nil {
		return err
	}
	if _, err = writer.Write([]byte{t}); err != nil {
		return err
	}
	body := base64.StdEncoding.EncodeToString(packet.Data)
	_, err = writer.Write([]byte(body))
	return err
}

func (p *b64Codec) encode(packet *Packet) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := p.writeTo(bf, packet); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func convertCharToType(c byte) (PacketType, error) {
	switch c {
	default:
		return 0xFF, fmt.Errorf("invalid packet type: %s", c)
	case '0':
		return OPEN, nil
	case '1':
		return CLOSE, nil
	case '2':
		return PING, nil
	case '3':
		return PONG, nil
	case '4':
		return MESSAGE, nil
	case '5':
		return UPGRADE, nil
	case '6':
		return NOOP, nil
	}
}

func convertTypeToChar(ptype PacketType) (byte, error) {
	switch ptype {
	default:
		return 0, fmt.Errorf("invalid packet type: %d", ptype)
	case OPEN:
		return '0', nil
	case CLOSE:
		return '1', nil
	case PING:
		return '2', nil
	case PONG:
		return '3', nil
	case MESSAGE:
		return '4', nil
	case UPGRADE:
		return '5', nil
	case NOOP:
		return '6', nil
	}
}
