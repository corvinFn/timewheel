package timewheel

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestTimeWheelDelay(t *testing.T) {
	t.Run("定时器延时执行回调", func(t *testing.T) {
		const delay = 30 * time.Millisecond
		num := int32(100)
		testChan := make(chan struct{})
		tw, _ := NewTimeWheel(10*time.Millisecond, 2)
		tw.activate()
		defer tw.stop()

		addNum := int32(100)
		startsAt := time.Now()

		tfun := func() {
			num += addNum
			duration := time.Since(startsAt)

			if duration < delay {
				println(fmt.Sprintf("Delay(%dms) actually-delay(%dms)", delay, duration))
				panic(errors.New("delay time wrong"))
			}
			if num != 200 {
				println(fmt.Sprintf("Expected(%d) actually(%d)", 200, num))
				panic(errors.New("delay func not exc"))
			}
			testChan <- struct{}{}
		}
		tw.addTimer(delay, tfun, false)

		<-testChan
	})
}

func TestTimeWheelRemove(t *testing.T) {
	t.Run("定时器删除", func(t *testing.T) {
		const delay = 5 * time.Millisecond
		num := int32(100)
		testChan := make(chan struct{})
		tw, _ := NewTimeWheel(10*time.Millisecond, 3600)
		tw.activate()
		defer tw.stop()

		addNum := int32(100)

		tfun := func() {
			num += addNum
			if num != 100 {
				println(fmt.Sprintf("Expected(%d) actually(%d)", 100, num))
				panic(errors.New("delay func not exc"))
			}
			testChan <- struct{}{}
		}
		wfun := func() {
			testChan <- struct{}{}
		}
		td1 := tw.addTimer(delay, tfun, false)
		td2 := tw.addTimer(delay, tfun, false)
		td3 := tw.addTimer(delay, tfun, false)
		tw.removeTimer(td1.Key)
		tw.removeTimer(td2.Key)
		tw.removeTimer(td3.Key)
		tw.addTimer(delay, wfun, false)

		<-testChan

	})
}

func TestTimeWheelRepeat(t *testing.T) {
	t.Run("定时器循环", func(t *testing.T) {
		const delay = 60 * time.Millisecond
		testChan := make(chan struct{})
		num := int32(100)
		repeatTime := int32(10)
		tw, _ := NewTimeWheel(10*time.Millisecond, 3600)
		tw.activate()
		defer tw.stop()

		tfun := func() {
			atomic.AddInt32(&num, 100)
		}
		tw.addTimer(delay, tfun, true)

		wfun := func() {
			if num != (repeatTime+1)*100 {
				panic(errors.New("timer repeat wrong"))
			}
			testChan <- struct{}{}
		}
		tw.addTimer(10*delay+10*time.Millisecond, wfun, false)

		<-testChan
	})
}

func TestTimeWheelReset(t *testing.T) {
	t.Run("定时器重置", func(t *testing.T) {
		const delay = 60 * time.Millisecond
		const sleepTime = 30 * time.Millisecond
		num := int32(100)
		testChan := make(chan struct{})
		tw, _ := NewTimeWheel(10*time.Millisecond, 3600)
		tw.activate()
		defer tw.stop()

		addNum := int32(100)
		startsAt := time.Now()

		f := func() {
			num += addNum

			duration := time.Since(startsAt)

			if duration < delay+sleepTime {
				println(fmt.Sprintf("Delay(%dms) actually-delay(%dms)", (delay+sleepTime)/1e6, duration/1e6))
				panic(errors.New("delay time wrong"))
			}
			if num != 200 {
				println(fmt.Sprintf("Expected(%d) actually(%d)", 200, num))
				panic(errors.New("delay func not exc"))
			}
			testChan <- struct{}{}
		}
		td := tw.addTimer(delay, f, false)
		td.Remove()
		time.Sleep(sleepTime)
		td.Restart()

		<-testChan
	})
}
