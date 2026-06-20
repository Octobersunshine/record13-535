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
