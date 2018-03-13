package promise

import "sync"

type Promise struct {
	done  bool
	sync  *sync.Cond
	Value interface{}
	Err   error
}

type ResolveFunc func(value interface{})
type RejectFunc func(err error)

func New(body func(resolve ResolveFunc, reject RejectFunc)) *Promise {

	p := &Promise{
		sync: sync.NewCond(&sync.Mutex{}),
	}
	go func() {
		body(func(value interface{}) {
			p.resolve(value)
		}, func(err error) {
			p.reject(err)
		})
	}()

	return p
}

func Resolve(value interface{}) *Promise {
	return &Promise{
		done:  true,
		Value: value,
	}
}

func Reject(err error) *Promise {
	return &Promise{
		done: true,
		Err:  err,
	}
}

func (p *Promise) resolve(value interface{}) {
	p.sync.L.Lock()
	defer p.sync.L.Unlock()
	if p.done {
		return
	}
	p.Value = value
	p.Err = nil
	p.done = true
	p.sync.Broadcast()
}

func (p *Promise) reject(err error) {
	p.sync.L.Lock()
	defer p.sync.L.Unlock()
	if p.done {
		return
	}
	p.Value = nil
	p.Err = err
	p.done = true
	p.sync.Broadcast()
}

func (p *Promise) Await() interface{} {
	if p.done {
		return p.Value
	}
	p.sync.L.Lock()
	for !p.done {
		p.sync.Wait()
	}
	p.sync.L.Unlock()
	return p.Value
}

func (p *Promise) Then(then ResolveFunc) *Promise {

	go func() {
		p.Await()
		if p.Err == nil {
			then(p.Value)
		}
	}()

	return p
}

func (p *Promise) Catch(catch RejectFunc) *Promise {

	go func() {
		p.Await()
		if p.Err != nil {
			catch(p.Err)
		}
	}()

	return p
}
