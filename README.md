# go-issue-24993

This repository is a reduced repro case for https://github.com/golang/go/issues/24993.

But, I've found that this seems to also trigger https://github.com/golang/go/issues/26243.

Just use go tip and `go run main.go`.

Notes:

- I've bisected the offending Go commit to [`9a8372f8b`](https://github.com/golang/go/commit/9a8372f8bd5a39d2476bfa9247407b51f9193b9e), although at that commit I only see the `sweep increased allocation count` error, not the `bad pointer` error.
- I have not been able to get the repro case to fail when running with the race detector.
- I have only attempted to repro on darwin/amd64, so I don't know if this occurs on any other OS/arch.

Sometimes it panics with a stack trace like

```
runtime: pointer 0xc000410000 to unallocated span span.base()=0xc000410000 span.limit=0xc000412000 span.state=3
runtime: found in object at *(0xc0002da540+0x10)
object=0xc0002da540 s.base()=0xc0002da000 s.limit=0xc0002dbfe0 s.spanclass=14 s.elemsize=96 s.state=mSpanInUse
 *(object+0) = 0x0
 *(object+8) = 0x10
 *(object+16) = 0xc000410000 <==
 *(object+24) = 0x100000010
 *(object+32) = 0x16d7800
 *(object+40) = 0x0
 *(object+48) = 0x0
 *(object+56) = 0x0
 *(object+64) = 0x0
 *(object+72) = 0x0
 *(object+80) = 0x0
 *(object+88) = 0x0
fatal error: found bad pointer in Go heap (incorrect use of unsafe or cgo?)

runtime stack:
runtime.throw(0x17a481e, 0x3e)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/panic.go:617 +0x72 fp=0x70000fad3d08 sp=0x70000fad3cd8 pc=0x102d892
runtime.findObject(0xc000410000, 0xc0002da540, 0x10, 0xd150, 0x2812800, 0xc000049c70)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mbitmap.go:397 +0x3b4 fp=0x70000fad3d58 sp=0x70000fad3d08 pc=0x1015c44
runtime.scanobject(0xc0002da540, 0xc000049c70)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mgcmark.go:1161 +0x216 fp=0x70000fad3de8 sp=0x70000fad3d58 pc=0x1021326
runtime.gcDrain(0xc000049c70, 0x3)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mgcmark.go:919 +0x217 fp=0x70000fad3e40 sp=0x70000fad3de8 pc=0x1020b37
runtime.gcBgMarkWorker.func2()
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mgc.go:1873 +0x80 fp=0x70000fad3e80 sp=0x70000fad3e40 pc=0x1054f20
runtime.systemstack(0x4c00000)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/asm_amd64.s:351 +0x66 fp=0x70000fad3e88 sp=0x70000fad3e80 pc=0x1056ee6
runtime.mstart()
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/proc.go:1153 fp=0x70000fad3e90 sp=0x70000fad3e88 pc=0x1031db0

...
```

And other times it fails with a stack trace like:

```
runtime: nelems=85 nalloc=14 previous allocCount=13 nfreed=65535
fatal error: sweep increased allocation count

goroutine 3 [running]:
runtime.throw(0x179837f, 0x20)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/panic.go:617 +0x72 fp=0xc000057668 sp=0xc000057638 pc=0x102d892
runtime.(*mspan).sweep(0x2683010, 0xc000080000, 0x1054400)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mgcsweep.go:326 +0x84b fp=0xc000057740 sp=0xc000057668 pc=0x10236fb
runtime.sweepone(0x17acc00)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mgcsweep.go:136 +0x26d fp=0xc0000577a8 sp=0xc000057740 pc=0x1022c5d
runtime.bgsweep(0xc000080000)
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mgcsweep.go:73 +0xb8 fp=0xc0000577d8 sp=0xc0000577a8 pc=0x1022948
runtime.goexit()
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/asm_amd64.s:1337 +0x1 fp=0xc0000577e0 sp=0xc0000577d8 pc=0x1058f21
created by runtime.gcenable
	/Users/mr/gotip/src/github.com/golang/go/src/runtime/mgc.go:208 +0x58

...
```

I've reproduced this with `go version devel +49abcf1a97 Thu Dec 20 09:04:35 2018 +0000 darwin/amd64`.
