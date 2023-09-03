# cake

`cake` is a simple package that uses Go generics, type embedding and the `reflect` package to provide a powerful pattern for layering Go interface implementations.

Another great way to think about `cake` is as middleware for your Go interfaces, as cake was inspired by [`net/http` middleware](https://pkg.go.dev/net/http#Handler).

**Goals**
- Promote the usage of the [Layered Architecture](https://en.wikipedia.org/wiki/Multitier_architecture) pattern in Go
- Enable easier [Separation of Concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) within any Go codebase

## Installation

```bash
go get github.com/tylermmorton/cake
```

## Usage

### Base

Every cake has a base. The base is the first layer of the cake and is the only layer that is required. The base is a struct that completely implements the interface that you would like to add layers to.

```go
package main

import "context"

// Service is a hypothetical chat service
type Service interface {
    GetMessage(ctx context.Context, id string) string
    CreateMessage(ctx context.Context, msg string) error
}

type baseLayer struct {
    // ...	
}

func (b *baseLayer) GetMessage(ctx context.Context, id string) string {
    // ...
}

func (b *baseLayer) CreateMessage(ctx context.Context, msg string) error {
    // ...
}
```

Once you have a base layer established you can start constructing your cake:
```go
package main

import (
    "context"

    "github.com/tylermmorton/cake"
)

func NewService() (Service, error) {
    return cake.Layered[Service](&baseLayer{})
}
```

Cakes with only a base layer will function as a very thin wrapper and will have no difference from the base layer itself. But really, what is that exciting about a cake with only one layer? The real power of `cake` comes from adding additional layers to your interface. 

### Layers

Layers are structs that implement the same interface type as the base layer [by embedding it](https://go101.org/article/type-embedding.html). The value of the embedded interface will be set dynamically to the next layer when the cake is being constructed. If there is no "next layer," cake will set the value of the embedded interface to the base layer.

```go
type loggingLayer struct {
    Service // <- embed the interface type of the cake, this will get set dynamically
}

func (l *loggingLayer) GetMessage(ctx context.Context, id string) string {
    log.Printf("GetMessage %s", id)
    return l.Service.GetMessage(ctx, id) // <- call the next layer, in this case: 'baseLayer'
}
```

And don't forget to add the layer to your cake:
```go
func NewService() (Service, error) {
    return cake.Layered[Service](
        &baseLayer{}, 
        &loggingLayer{}, 
    )
}
```
#### Fallthroughs

You might be thinking: The `loggingLayer` in the example above does not implement the `CreateMessage` method! And you would be correct. When a method is called on a layer that doesn't implement it, cake will _fallthrough_ to the next layer that has a valid implementation. And again, if there is no "next" layer, cake will fallthrough all the way to the base layer.

## Strategies

### Deferred work
Just like with `net/http` middleware, a powerful feature of layered architecture is the ability to defer work until other layers have completed. This is especially useful for things like logging, metrics, tracing, etc.

Below, the `loggingLayer` is able to defer its work until _after_ all other layers are done, so it can use the result of the call in its log statement.

```go
type loggingLayer struct {
    Service
}

func (l *loggingLayer) GetMessage(ctx context.Context, id string) string {
    msg := l.Service.GetMessage(ctx, id) // <- call the next layer, in this case: 'baseLayer'
	
    // After other layers have completed their work, it's time to log the result
    log.Printf("GetMessage %s: %q", id. msg)
	
    return msg
}
```

### Conditional work
It may be useful to execute some work conditionally. For example, you may want to be able to enable/disable the logging system in your application.

```go
type loggingLayer struct {
    Service
}

func (l *loggingLayer) GetMessage(ctx context.Context, id string) string {
    if os.Getenv("LOGGING_ENABLED") != "true" {
        return l.Service.GetMessage(ctx, id)
    }
	
    msg := l.Service.GetMessage(ctx, id)
	
    // After other layers have completed their work, it's time to log the result
    log.Printf("GetMessage %s: %q", id. msg)
	
    return msg
}
```

### Short-circuiting
Another powerful feature of layered architecture is the ability to short-circuit the call stack and return early. This is especially useful for things like authorization, validation, etc.

Below, this hypothetical `authLayer` is able to short-circuit the call stack of the cake and return early if the user is not authorized to perform the requested action.

```go
type authLayer struct {
    Service
}

func (l *authLayer) CreateMessage(ctx context.Context, msg string) error {
    if !l.authorized(ctx) {
        return ErrUnauthorized
    }
    
    return l.Service.CreateMessage(ctx, msg) // <- call the next layer, in this case: 'baseLayer'
}
```
