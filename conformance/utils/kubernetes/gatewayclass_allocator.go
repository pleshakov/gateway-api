package kubernetes

import "fmt"

// GatewayClassAllocator allocates a GatewayClass for Gateways.
type GatewayClassAllocator interface {
	Allocate() string
	Free(string)
}

// SingleGatewayClassAllocator always allocates the same GatewayClass.
// The GatewayClass is always allocated no matter how many times Allocate() is called.
type SingleGatewayClassAllocator struct {
	gatewayClass string
}

func NewSingleGatewayClassAllocator(gatewayClass string) *SingleGatewayClassAllocator {
	return &SingleGatewayClassAllocator{
		gatewayClass: gatewayClass,
	}
}

func (a *SingleGatewayClassAllocator) Allocate() string {
	return a.gatewayClass
}

func (a *SingleGatewayClassAllocator) Free(string) {
	// no action necessary
}

// MultiGatewayClassAllocator allocates a GatewayClass from a list of GatewayClasses.
// An allocation makes the allocated GatewayClass unavailable for further allocations until it is freed.
type MultiGatewayClassAllocator struct {
	gatewayClasses []string
	allocations    map[string]struct{}
}

func NewMultiGatewayClassAllocator(gatewayClasses []string) *MultiGatewayClassAllocator {
	return &MultiGatewayClassAllocator{
		gatewayClasses: gatewayClasses,
		allocations:    make(map[string]struct{}),
	}
}

func (a *MultiGatewayClassAllocator) Allocate() string {
	for _, gc := range a.gatewayClasses {
		if _, exist := a.allocations[gc]; exist {
			continue
		}

		a.allocations[gc] = struct{}{}
		return gc
	}

	panic("no GatewayClass available")
}

func (a *MultiGatewayClassAllocator) Free(gc string) {
	if _, exist := a.allocations[gc]; !exist {
		panic(fmt.Sprintf("GatewayClass %s is not allocated", gc))
	}

	delete(a.allocations, gc)
}
