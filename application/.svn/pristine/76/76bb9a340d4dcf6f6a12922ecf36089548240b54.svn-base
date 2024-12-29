package utils

import (
	"application/pkg/utils/log"
	"fmt"
	"runtime"
)

func SafeGo(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("发生了panic错误:", err)
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				fmt.Printf("发生错误，堆栈信息如下：\n%s\n", buf[:n])
			}
		}()
		f()
	}()
}

func LoopSafeGo(f func()) {
	go func() {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Error("发生了panic错误:", err)
						buf := make([]byte, 4096)
						n := runtime.Stack(buf, false)
						fmt.Printf("发生错误，堆栈信息如下：\n%s\n", buf[:n])
					}
				}()
				f()
			}()
		}
	}()
}
