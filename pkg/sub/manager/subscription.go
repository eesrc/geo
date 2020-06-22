package manager

// Subscription ...
type Subscription interface {
	Unsubscribe() error
	GetChan() <-chan interface{}
}
