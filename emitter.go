package SocketIO

import (
	"bytes"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/vmihailenco/msgpack"
	"strconv"
)

const (
	EVENT        = 2
	BINARY_EVENT = 5
)

type EmitterOpts struct {
	// Host means hostname like localhost
	Host string
	// Port means port number, like 6379
	Port int
	// Key means redis subscribe key
	Key string
	// Protocol, like tcp
	Protocol string
	// Address, like localhost:6379
	Addr string
}

type Emitter struct {
	Redis     redis.Conn
	Prefix    string
	Namespace string
	Channel   string
	rooms     []string
	flags     map[string]interface{}
}

// Emitter constructor
// Usage:
// SocketIO.NewEmitter(&SocketIO.EmitterOpts{
//    Host:"localhost",
//    Port:6379,
// })
func NewEmitter(opts *EmitterOpts) (*Emitter, error) {
	var addr string
	if opts.Addr != "" {
		addr = opts.Addr
	} else if opts.Host != "" && opts.Port > 0 {
		addr = opts.Host + ":" + strconv.Itoa(opts.Port)
	} else {
		addr = "localhost:6379"
	}
	var protocol string
	if opts.Protocol == "" {
		protocol = "tcp"
	} else {
		protocol = opts.Protocol
	}
	conn, err := redis.Dial(protocol, addr)
	if err != nil {
		return nil, err
	}

	var prefix string
	if opts.Key == "" {
		prefix = "socket.io"
	} else {
		prefix = opts.Key
	}

	// Default namespace.
	nsp := "/"

	// Build channel name compatible with: https://github.com/socketio/socket.io-emitter
	delimiter := "#"
	channel := fmt.Sprintf("%s%s%s%s", prefix, delimiter, nsp, delimiter)

	emitter := &Emitter{
		Redis:     conn,
		Prefix:    prefix,
		Namespace: nsp,
		Channel:   channel,
	}
	return emitter, nil
}

func (emitter *Emitter) Join() *Emitter {
	emitter.flags["join"] = true
	return emitter
}

func (emitter *Emitter) Volatile() *Emitter {
	emitter.flags["volatile"] = true
	return emitter
}

func (emitter *Emitter) Broadcast() *Emitter {
	emitter.flags["broadcast"] = true
	return emitter
}

/**
 * Limit emission to a certain `room`.
 *
 * @param {String} room
 */
func (emitter *Emitter) In(room string) *Emitter {
	for _, r := range emitter.rooms {
		if r == room {
			return emitter
		}
	}
	emitter.rooms = append(emitter.rooms, room)
	return emitter
}

func (emitter *Emitter) To(room string) *Emitter {
	return emitter.In(room)
}

/**
 * Limit emission to certain `namespace`.
 *
 * @param {String} namespace
 */
func (emitter *Emitter) Of(namespace string) *Emitter {
	emitter.Namespace = namespace
	return emitter
}

// send the packet by string, json, etc
// Usage:
// Emit("event name", "data")
func (emitter *Emitter) Emit(event string, data ...interface{}) (*Emitter, error) {
	// Build data from args.
	d := []interface{}{event}
	d = append(d, data...)

	// Determine event type.
	eventType := EVENT
	if HasBinary(data...) {
		eventType = BINARY_EVENT
	}

	// Build packet.
	packet := map[string]interface{}{
		"type": eventType,
		"data": d,
		"nsp":  emitter.Namespace,
	}

	return emitter.emit(packet)
}

// send the packet by binary
// Usage:
// EmitBinary("event name", []byte{0x01, 0x02, 0x03})
func (emitter *Emitter) EmitBinary(event string, data ...interface{}) (*Emitter, error) {
	// Build data from args.
	d := []interface{}{event}
	d = append(d, data...)

	// Build packet.
	packet := map[string]interface{}{
		"type": BINARY_EVENT,
		"data": d,
		"nsp":  emitter.Namespace,
	}
	return emitter.emit(packet)
}

func HasBinary(dataSlice ...interface{}) bool {
	if dataSlice == nil {
		return false
	}
	for _, data := range dataSlice {
		switch dataType := data.(type) {
		case []byte:
			return true
		case bytes.Buffer:
			return true
		case []interface{}:
			for _, d := range dataType {
				result := HasBinary(d)
				if result {
					return true
				}
			}
		case map[string]interface{}:
			for _, v := range dataType {
				result := HasBinary(v)
				if result {
					return true
				}
			}
		default:
			return false
		}
	}
	return false
}

func (emitter *Emitter) emit(packet map[string]interface{}) (*Emitter, error) {

	// Channel delimiter.
	delimiter := "#"

	// Build current channel name.
	channel := fmt.Sprintf("%s%s%s%s", emitter.Prefix, delimiter, emitter.Namespace, delimiter)

	// Add room to channel.
	if len(emitter.rooms) == 1 {
		channel = fmt.Sprintf("%s%s%s", channel, emitter.rooms[0], "#")
	}

	// Build packet.
	var pack []interface{} = make([]interface{}, 0)

	// Add UID to packet.
	pack = append(pack, "emitter")

	// Add packet data.
	pack = append(pack, packet)

	// Add options.
	pack = append(pack, map[string]interface{}{
		"rooms": emitter.rooms,
		"flags": emitter.flags,
	})

	// Encode.
	buf := &bytes.Buffer{}
	enc := msgpack.NewEncoder(buf)
	error := enc.Encode(pack)

	// Handle encoder error.
	if error != nil {
		return nil, error
	}

	// Publish to Redis.
	emitter.Redis.Do("PUBLISH", channel, buf)

	// Reset state.
	emitter.rooms = []string{}
	emitter.flags = make(map[string]interface{})

	return emitter, nil
}

func (emitter *Emitter) Close() {
	if emitter.Redis != nil {
		defer emitter.Redis.Close()
	}
}
