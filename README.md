socket.io-go-emitter
=====================

[![Build Status](https://travis-ci.org/yosuke-furukawa/socket.io-go-emitter.svg?branch=master)](https://travis-ci.org/yosuke-furukawa/socket.io-go-emitter)

A Golang implementation of [socket.io-emitter](https://github.com/Automattic/socket.io-emitter)

This project was forked to make the Redis PUBLISH channel compatible with the Node.js implementation of the Socket.IO Redis adapter: [socket.io-redis](https://github.com/socketio/socket.io-redis).

This project uses redis.
Make sure your environment has redis.

Install and development
--------------------

To install in your golang project.

```sh
$ go get github.com/yosuke-furukawa/socket.io-go-emitter
```

Usage
---------------------

Example:

```go
  emitter, _ := SocketIO.NewEmitter(&SocketIO.EmitterOpts{
    Host:"localhost",
    Port:6379,
  })
  emitter.Emit("message", "I love you!!")
```

### Broadcasting and other flags

Possible flags

- json
- volatile
- broadcast

```go
  emitter, _ := SocketIO.NewEmitter(&SocketIO.EmitterOpts{
    Host:"localhost",
    Port:6379,
  })
  emitter.Volatile().Emit("message", "I love you!!")
```

** Binary Support

```go
  emitter, _ := SocketIO.NewEmitter(&SocketIO.EmitterOpts{
    Host:"localhost",
    Port:6379,
  })
  val := bytes.NewBufferString("I love you!!")
  emitter.EmitBinary("message", val)
```

