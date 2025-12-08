package muxval

import "sync"

type MuxVal[T any] struct {
	mu  sync.Mutex
	val T
}

func (mv *MuxVal[T]) Modify(f func(T) T) {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	mv.val = f(mv.val)
}

func (mv *MuxVal[T]) Read(f func(T)) {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	f(mv.val)
}
