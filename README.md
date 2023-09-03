# cake

`cake` is a package that uses Go generics, type embedding and the `reflect` package together to provide a powerful pattern for layering Go interface implementations.

Another great way to think about `cake` layering is like middleware for your Go interfaces, as cake was inspired by [`net/http` middlewares](https://pkg.go.dev/net/http#Handler).

**Goals**
- Promote the usage of the [Layered Architecture](https://en.wikipedia.org/wiki/Multitier_architecture) pattern in Go
- Enable easier [Separation of Concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) within any Go codebase

## Installation

```bash
go get github.com/tylermmorton/cake
```

## Usage

### Base

Every layered cake must start with a base. The base is the first layer of the cake and it supports all of the other layers, therefore it is required. In code, the base is a struct that implements the interface that you would like to add layers to. Chances are you already have an interface implementation that is suitable as a base layer.

```go
package main

import "context"

// Service is a hypothetical chat service that is able to get/set messages in a database
type Service interface {
    GetMessage(ctx context.Context, id string) string
    CreateMessage(ctx context.Context, msg string) error
}

type baseLayer struct {
    // ...	
}

func (l *baseLayer) GetMessage(ctx context.Context, id string) string {
    return l.db.GetMessage(ctx, id)
}

func (l *baseLayer) CreateMessage(ctx context.Context, msg string) error {
    return l.db.CreateMessage(ctx, msg)
}
```

### Construction

Once you have a base established you can start constructing your layered cake by simply calling the generic function `Layered`:

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

The signature of `Layered` is as follows:

```go
func Layered[T interface{}](base T, layers ...T) (T, error)
```

`Layered` takes a `base` layer and any number of additional layers to add. 

Providing just a base for a cake will still work. But really, what is exciting about a cake with only one layer? The real power of `cake` comes from adding additional layers to your interface. 

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
        &baseLayer{}, // <- base
        &loggingLayer{}, // <- layer[0]
    )
}
```

Now, when calling `GetMessage` on our layered `Service`, each layer will be called in the order it was provided, with the base layer being the final layer in the call stack.

### Fallthroughs

Those with a keen eye will notice that the `loggingLayer` in the example above does not implement the `CreateMessage` method! When a method is called on a layer that doesn't implement it, cake will _fallthrough_ to the "next layer" that has a valid implementation. And again, if there is no "next layer", cake will fallthrough all the way to the base layer.

## Patterns

Below are some useful patterns that can be leveraged in an application built with a layered architecture.

### Yield

Just like with `net/http` middleware, a powerful feature of layered architecture is the ability to yield until other layers have completed their work. This is especially useful for layering auxillary systems such as logging, metrics, tracing, etc. This could also be known in some systems as 'deferred work' or 'exitware.'

Below is a modified version of the `loggingLayer` from earlier, but now it calls into the next layer _before_ logging any output. This allows you to wait until other layers have finished their work and returned a result before printing a log, making the log statement much more detailed and useful.

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
    // Call into the next layer and skip logging if its not enabled.
    if os.Getenv("LOGGING_ENABLED") != "true" {
        return l.Service.GetMessage(ctx, id)
    }
	
    msg := l.Service.GetMessage(ctx, id)
	
    log.Printf("GetMessage %s: %q", id. msg)
	
    return msg
}
```

### Short-circuiting

Another powerful feature of layered architecture is the ability to short-circuit the call stack and return early. This is especially useful for things like authorization, validation, memoization, caching, etc.

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
