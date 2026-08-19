# Bug 复现说明

## Bug 是什么
并发取件时，`PickupService.Pickup` 先在仓储锁外读取 locker 和 parcel，再调用 `Store.ReleaseLocker`。另一个取件请求可以在这段间隙修改同一 locker，导致并发读写同一对象并产生数据竞争。

## 如何触发
启动含 Bug 的后端代码后，在同一个已占用格口上并发发起多个取件请求；验证测试会用多个 goroutine 同时调用 `PickupService.Pickup`。

## 运行指令
```bash
cd /workplace/kd-payment-system__001/backend
go test ./internal/service -race -run '^TestConcurrentPickupSingleCommit$' -count=1
```

## 错误信息
Go race detector 报告同一个 locker 字段被并发读写，测试以非零状态退出。

## 错误堆栈
```text
WARNING: DATA RACE
Write at 0x00c0000152a0 by goroutine 20:
  kd-payment-system/backend/internal/store.(*Store).ReleaseLocker()
      backend/internal/store/mutations.go:70 +0x178
  kd-payment-system/backend/internal/service.(*PickupService).Pickup()
      backend/internal/service/pickup_svc.go:41 +0xec
  kd-payment-system/backend/internal/service_test.TestConcurrentPickupSingleCommit.func1()
      backend/internal/service/concurrent_pickup_test.go:53 +0x9c
Previous read at 0x00c0000152a0 by goroutine 43:
  kd-payment-system/backend/internal/service.(*PickupService).Pickup()
      backend/internal/service/pickup_svc.go:33 +0x6c
  kd-payment-system/backend/internal/service_test.TestConcurrentPickupSingleCommit.func1()
      backend/internal/service/concurrent_pickup_test.go:53 +0x9c
--- FAIL: TestConcurrentPickupSingleCommit
    testing.go:1712: race detected during execution of test
FAIL
```

## 稳定性校准
验证资产使用统一启动 channel 和 WaitGroup 让多个取件 goroutine 同时进入临界路径；单次 `-count=1` 已能命中 race。额外使用以下命令重复 10 次进行稳定性校准，含 Bug 环境每次均由 `-race` 报告竞态：

```bash
cd /workplace/kd-payment-system__001/backend
go test ./internal/service -race -run '^TestConcurrentPickupSingleCommit$' -count=10
```
