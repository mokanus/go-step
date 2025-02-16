package g

import (
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/log/ubiquitous/log/field"
	"runtime/debug"
)

func Go(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.GetLogger().Error("goroutine panic", field.String("stack", string(debug.Stack())))
			}
		}()

		f()
	}()
}
