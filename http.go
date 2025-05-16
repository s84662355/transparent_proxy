package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"transparent/log"
)

func gohttp() {
	runtime.SetMutexProfileFraction(10) // 开启对锁调用的跟踪
	runtime.SetBlockProfileRate(10)     // 开启对阻塞操作的跟踪
	runtime.MemProfileRate = 128 * 1024
	go func() {
		defer log.Recover("http")
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}
		go func() {
			for {
				log.Debug(fmt.Sprint("端口:", ln.Addr().(*net.TCPAddr).Port))
				log.Debug(fmt.Sprint("metrics: ", fmt.Sprintf("http://127.0.0.1:%d/metrics", ln.Addr().(*net.TCPAddr).Port)))
				log.Debug(fmt.Sprint("pprof: ", fmt.Sprintf("http://127.0.0.1:%d/debug/pprof", ln.Addr().(*net.TCPAddr).Port)))
				<-time.After(30 * 60 * time.Second)
			}
		}()

		http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			result := fmt.Sprint("NumGoroutine:", runtime.NumGoroutine(), " 当前协程数\n")                          ///当前协程数
			result += fmt.Sprint("memStats.Alloc:", memStats.Alloc, "  目前由 Go 程序分配的字节数，不包括由垃圾回收器管理的内存。\n")     // 目前由 Go 程序分配的字节数，不包括由垃圾回收器管理的内存。
			result += fmt.Sprint("memStats.TotalAlloc:", memStats.TotalAlloc, " 自程序启动以来分配的总字节数，包括已经释放的内存。\n")  // 自程序启动以来分配的总字节数，包括已经释放的内存。
			result += fmt.Sprint("memStats.Sys:", memStats.Sys, " 总共从操作系统获得的内存字节数，包括已经被释放回系统的内存\n")            // 总共从操作系统获得的内存字节数，包括已经被释放回系统的内存。
			result += fmt.Sprint("memStats.Mallocs:", memStats.Mallocs, " 总共进行的内存分配次数。\n")                     // 总共进行的内存分配次数。
			result += fmt.Sprint("memStats.Frees:", memStats.Frees, " 总共进行的内存释放次数。\n")                         // 总共进行的内存释放次数。
			result += fmt.Sprint("memStats.HeapAlloc:", memStats.HeapAlloc, " 目前在堆上分配的字节数。\n")                 // 目前在堆上分配的字节数。
			result += fmt.Sprint("memStats.HeapSys:", memStats.HeapSys, " 总共从操作系统获得的堆内存字节数。\n")                // 总共从操作系统获得的堆内存字节数。
			result += fmt.Sprint("memStats.HeapIdle:", memStats.HeapIdle, " 目前未被使用，但已从系统保留的堆内存字节数。\n")         // 目前未被使用，但已从系统保留的堆内存字节数。
			result += fmt.Sprint("memStats.HeapInuse:", memStats.HeapInuse, " 目前在堆上使用的内存字节数。\n")               // 目前在堆上使用的内存字节数。
			result += fmt.Sprint("memStats.HeapReleased:", memStats.HeapReleased, " 已经返回给操作系统的堆内存字节数。\n")      // 已经返回给操作系统的堆内存字节数。
			result += fmt.Sprint("memStats.HeapObjects:", memStats.HeapObjects, " 目前在堆上的对象数。\n")               // 目前在堆上的对象数。
			result += fmt.Sprint("memStats.StackInuse:", memStats.StackInuse, " 目前在栈上使用的内存字节数。\n")             // 目前在栈上使用的内存字节数。
			result += fmt.Sprint("memStats.StackSys:", memStats.StackSys, " 为栈操作从操作系统获得的内存字节数。\n")             // 为栈操作从操作系统获得的内存字节数。
			result += fmt.Sprint("memStats.MSpanInuse:", memStats.MSpanInuse, " 目前在 MSpan 结构体上使用的内存字节数。\n")    ///目前在 MSpan 结构体上使用的内存字节数。
			result += fmt.Sprint("memStats.MSpanSys:", memStats.MSpanSys, " 为 MSpan 结构体从操作系统获得的内存字节数。\n")      // 为 MSpan 结构体从操作系统获得的内存字节数。
			result += fmt.Sprint("memStats.MCacheInuse:", memStats.MCacheInuse, " 目前在 MCache 结构体上使用的内存字节数。\n") ///目前在 MCache 结构体上使用的内存字节数。
			result += fmt.Sprint("memStats.MCacheSys:", memStats.MCacheSys, " 	为 MCache 结构体从操作系统获得的内存字节数。\n")  // MCacheSys: 为 MCache 结构体从操作系统获得的内存字节数。
			result += fmt.Sprint("memStats.BuckHashSys:", memStats.BuckHashSys, " 为桶哈希表从操作系统获得的内存字节数。\n")      ///BuckHashSys: 为桶哈希表从操作系统获得的内存字节数。
			result += fmt.Sprint("memStats.GCSys:", memStats.GCSys, " 为垃圾回收器从操作系统获得的内存字节数。\n")                 /// 为垃圾回收器从操作系统获得的内存字节数。
			result += fmt.Sprint("memStats.OtherSys:", memStats.OtherSys, " 为其他内存管理用途从操作系统获得的内存字节数。\n")        ///  为其他内存管理用途从操作系统获得的内存字节数。

			fmt.Fprintf(w, result)
		})

		panic(http.Serve(ln, nil))
	}()
}
