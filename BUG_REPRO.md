# Bug 复现说明

## Bug 是什么

当请求尺寸在目标柜机及其邻近柜机都没有可用格口时，Scheduler.Decide 只返回未占用结果而没有返回错误。DropoffService 因而把容量耗尽误报成 err=nil，调用方无法区分正常未占用与容量不足。

## 如何触发

使用测试构造的无可用容量区域，调用 DropoffService.Dropoff 申请指定尺寸的包裹，并检查返回值是否包含容量错误。

## 运行指令

在项目根目录（验证环境的 `env` 目录）执行：

```bash
go test ./internal/service -run '^TestDropoffReportsCapacityWhenNoCabinetCanAcceptParcel$' -count=1
```

## 错误信息

该命令在含 Bug 的初始代码上失败。关键错误信息为：

`expected capacity error`，实际返回 `err=<nil>`。

## 错误堆栈

以下代码块保留该命令实际运行时的原始输出：

```text
--- FAIL: TestDropoffReportsCapacityWhenNoCabinetCanAcceptParcel (0.00s)
    dropoff_capacity_test.go:45: expected capacity error, got result={Schedule:{RequestedCabinetID:cab-cbd-02 Size:L RecommendedCabinetID:cab-cbd-02 RecommendedQuote:0x1d015566c980 DistanceMeters:0 IsRedirected:false Reason:目标柜机及邻近柜机均无可用格口，请稍后再试 Alternatives:[]} LockerID: ParcelID: PricePaid:0 Occupied:false} err=<nil>
FAIL
FAIL	kd-payment-system/backend/internal/service	0.857s
FAIL
```
