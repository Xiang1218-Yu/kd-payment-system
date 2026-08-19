# Bug 复现说明

## Bug 是什么

Seed 初始化只把预置格口标记为已占用，没有建立对应的 Parcel 和 ParcelID 关联。取件流程拿到空的包裹指针后，Store.ReleaseLocker 在生成历史记录时直接解引用，触发空指针 panic。

## 如何触发

使用 Seed 创建默认 Store 和 PickupService，然后按验证测试取件预置占用格口。该路径会进入 PickupService.Pickup 和 Store.ReleaseLocker。

## 运行指令

在项目根目录（验证环境的 `env` 目录）执行：

```bash
go test ./internal/service -run '^TestPickupSeededOccupiedLockerRecordsHistory$' -count=1
```

## 错误信息

该命令在含 Bug 的初始代码上失败。关键错误信息为：

`panic: runtime error: invalid memory address or nil pointer dereference`。

## 错误堆栈

以下代码块保留该命令实际运行时的原始输出：

```text
--- FAIL: TestPickupSeededOccupiedLockerRecordsHistory (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered, repanicked]
[signal SIGSEGV: segmentation violation code=0x2 addr=0x0 pc=0x1009c5910]

goroutine 22 [running]:
testing.tRunner.func1.2({0x100b03a20, 0x100b4cd10})
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:1974 +0x1a0
testing.tRunner.func1()
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:1977 +0x318
panic({0x100b03a20?, 0x100b4cd10?})
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/runtime/panic.go:860 +0x12c
kd-payment-system/backend/internal/store.(*Store).ReleaseLocker(0x386c6e4a02a0, {0x386c6e4a6610, 0xd}, 0x0)
	/Users/tog/Desktop/code/go标注/我的go/2026-08-19/kd-payment-system__003/env/internal/store/mutations.go:72 +0x110
kd-payment-system/backend/internal/service.(*PickupService).Pickup(0x386c6e4acf10, {0x386c6e4a6610, 0xd})
	/Users/tog/Desktop/code/go标注/我的go/2026-08-19/kd-payment-system__003/env/internal/service/pickup_svc.go:41 +0x84
kd-payment-system/backend/internal/service.TestPickupSeededOccupiedLockerRecordsHistory(0x386c6e50c248)
	/Users/tog/Desktop/code/go标注/我的go/2026-08-19/kd-payment-system__003/env/internal/service/seeded_pickup_test.go:34 +0x1fc
testing.tRunner(0x386c6e50c248, 0x100b2dfb8)
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:2036 +0xc4
created by testing.(*T).Run in goroutine 1
	/opt/homebrew/Cellar/go/1.26.5/libexec/src/testing/testing.go:2101 +0x3a8
FAIL	kd-payment-system/backend/internal/service	0.779s
FAIL
```
