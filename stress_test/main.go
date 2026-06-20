package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"ticket-reservation/lock"
	"ticket-reservation/model"
	"ticket-reservation/store"
	"time"
)

func main() {
	testConcurrencySafety()
	fmt.Println()
	testQuotaAdjustConcurrency()
	fmt.Println()
	testQuotaAdjustEdgeCases()
}

func testConcurrencySafety() {
	locker := lock.NewMemoryLocker()
	s := store.New(locker)

	slot := &model.TimeSlot{
		ID:        "slot-concurrent-test",
		Date:      "2026-07-01",
		StartTime: "09:00",
		EndTime:   "12:00",
		Total:     100,
		Reserved:  0,
		CreatedAt: time.Now(),
	}
	if err := s.CreateSlot(slot); err != nil {
		panic(err)
	}

	const (
		concurrentUsers = 1000
		perRequestQty   = 1
	)

	var (
		successCount int64
		failCount    int64
		wg           sync.WaitGroup
	)

	start := time.Now()

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			_, err := s.Reserve("slot-concurrent-test", perRequestQty, fmt.Sprintf("user-%d", userID))
			if err != nil {
				atomic.AddInt64(&failCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	finalSlot, _ := s.GetSlot("slot-concurrent-test")

	fmt.Println("=== 并发预约压测结果 ===")
	fmt.Printf("时段总量:           %d\n", slot.Total)
	fmt.Printf("并发请求数:         %d\n", concurrentUsers)
	fmt.Printf("每单请求数量:       %d\n", perRequestQty)
	fmt.Printf("成功预约:           %d 单\n", successCount)
	fmt.Printf("拒绝预约:           %d 单\n", failCount)
	fmt.Printf("成功预约总票数:     %d\n", successCount*int64(perRequestQty))
	fmt.Printf("时段最终已预约:     %d\n", finalSlot.Reserved)
	fmt.Printf("时段最终剩余:       %d\n", finalSlot.Remaining())
	fmt.Printf("耗时:               %v\n", duration)
	fmt.Println()

	if finalSlot.Reserved > slot.Total {
		fmt.Printf("❌ 超卖！已预约 %d > 总量 %d\n", finalSlot.Reserved, slot.Total)
	} else if successCount*int64(perRequestQty) != int64(finalSlot.Reserved) {
		fmt.Printf("❌ 数据不一致！成功预约 %d 但已预约 %d\n", successCount*int64(perRequestQty), finalSlot.Reserved)
	} else if finalSlot.Reserved != slot.Total {
		fmt.Printf("⚠️  未约满，已预约 %d，剩余 %d\n", finalSlot.Reserved, finalSlot.Remaining())
	} else {
		fmt.Println("✅ 测试通过！无超卖，数据一致")
	}
}

func testQuotaAdjustConcurrency() {
	locker := lock.NewMemoryLocker()
	s := store.New(locker)

	slot := &model.TimeSlot{
		ID:        "slot-quota-test",
		Date:      "2026-07-01",
		StartTime: "13:00",
		EndTime:   "16:00",
		Total:     50,
		Reserved:  0,
		CreatedAt: time.Now(),
	}
	if err := s.CreateSlot(slot); err != nil {
		panic(err)
	}

	const (
		reserveGoroutines = 200
		adjustGoroutines  = 100
		perReserveQty     = 1
	)

	var (
		reserveSuccess int64
		reserveFail    int64
		adjustSuccess  int64
		adjustFail     int64
		wg             sync.WaitGroup
	)

	start := time.Now()

	for i := 0; i < reserveGoroutines; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			_, err := s.Reserve("slot-quota-test", perReserveQty, fmt.Sprintf("user-%d", userID))
			if err != nil {
				atomic.AddInt64(&reserveFail, 1)
			} else {
				atomic.AddInt64(&reserveSuccess, 1)
			}
		}(i)
	}

	for i := 0; i < adjustGoroutines; i++ {
		wg.Add(1)
		go func(adjustID int) {
			defer wg.Done()
			delta := 1
			if adjustID%2 == 0 {
				delta = -1
			}
			_, err := s.AdjustQuota("slot-quota-test", delta)
			if err != nil {
				atomic.AddInt64(&adjustFail, 1)
			} else {
				atomic.AddInt64(&adjustSuccess, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	finalSlot, _ := s.GetSlot("slot-quota-test")

	fmt.Println("=== 预约+调额并发压测结果 ===")
	fmt.Printf("初始总量:           %d\n", slot.Total)
	fmt.Printf("预约协程数:         %d\n", reserveGoroutines)
	fmt.Printf("调额协程数:         %d\n", adjustGoroutines)
	fmt.Printf("预约成功:           %d 单\n", reserveSuccess)
	fmt.Printf("预约失败:           %d 单\n", reserveFail)
	fmt.Printf("调额成功:           %d 次\n", adjustSuccess)
	fmt.Printf("调额失败:           %d 次\n", adjustFail)
	fmt.Printf("时段最终总量:       %d\n", finalSlot.Total)
	fmt.Printf("时段最终已预约:     %d\n", finalSlot.Reserved)
	fmt.Printf("时段最终剩余:       %d\n", finalSlot.Remaining())
	fmt.Printf("耗时:               %v\n", duration)
	fmt.Println()

	if finalSlot.Reserved > finalSlot.Total {
		fmt.Printf("❌ 超卖！已预约 %d > 总量 %d\n", finalSlot.Reserved, finalSlot.Total)
	} else {
		fmt.Println("✅ 测试通过！无超卖，调额与预约并发安全")
	}
}

func testQuotaAdjustEdgeCases() {
	locker := lock.NewMemoryLocker()
	s := store.New(locker)

	slot := &model.TimeSlot{
		ID:        "slot-edge-test",
		Date:      "2026-07-01",
		StartTime: "17:00",
		EndTime:   "19:00",
		Total:     100,
		Reserved:  30,
		CreatedAt: time.Now(),
	}
	if err := s.CreateSlot(slot); err != nil {
		panic(err)
	}

	fmt.Println("=== 配额调整边界测试 ===")

	result, err := s.AdjustQuota("slot-edge-test", 50)
	if err != nil {
		fmt.Printf("❌ 增加配额失败: %v\n", err)
	} else {
		fmt.Printf("✅ 增加配额 +50: total=%d, reserved=%d\n", result.Total, result.Reserved)
	}

	result, err = s.AdjustQuota("slot-edge-test", -121)
	if err != nil {
		fmt.Printf("✅ 减到低于已预约被拦截: %v\n", err)
	} else {
		fmt.Printf("❌ 未拦截: total=%d, reserved=30\n", result.Total)
	}

	result, err = s.AdjustQuota("slot-nonexistent", 10)
	if err != nil {
		fmt.Printf("✅ 不存在的时段被拦截: %v\n", err)
	} else {
		fmt.Printf("❌ 不存在的时段未拦截: total=%d\n", result.Total)
	}
}
