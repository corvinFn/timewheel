package timewheel

import (
	"container/list"
	"errors"
	"time"

	"github.com/renstrom/shortuuid"
)

var (
	tw *TimeWheel
)

// CallBack task
type CallBack func()

// based on https://github.com/ouqiang/timewheel
type TimeWheel struct {
	noCopy noCopy

	interval          time.Duration // time interval to moving forward
	ticker            *time.Ticker
	slots             []*list.List
	timer             map[string]Record // unique key to task info
	curPos            int64             // current position
	slotNum           int64             // size of slots
	addTaskChannel    chan *Task        // channel to add task
	removeTaskChannel chan string       // channel to delete task
	stopChannel       chan chan struct{}
}

// Task 延时任务
type Task struct {
	delay    time.Duration
	circle   int64  // task excuted when circle == 0
	key      string // unique id
	repeat   bool   // circular task or not
	callback CallBack
}

type Record struct {
	pos  int64
	task *Task
}

type TimerData struct {
	timewheel *TimeWheel
	task      *Task
	Key       string
}

func InitTimeWheel(interval time.Duration, slotNum int64) error {
	var err error
	tw, err = NewTimeWheel(interval, slotNum)
	if err != nil {
		return err
	}
	tw.activate()

	return nil
}

// new creates a timewheel with specified interval and slots size
func NewTimeWheel(interval time.Duration, slotNum int64) (*TimeWheel, error) {
	if interval <= 0 || slotNum <= 0 {
		return nil, errors.New("args-invalid")
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		timer:             make(map[string]Record),
		curPos:            0,
		slotNum:           slotNum,
		addTaskChannel:    make(chan *Task),
		removeTaskChannel: make(chan string),
		stopChannel:       make(chan chan struct{}),
	}
	tw.initSlots()

	return tw, nil
}

// init every slot with empty list before activate
func (tw *TimeWheel) initSlots() {
	for i := int64(0); i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

// activate the timewheel
func (tw *TimeWheel) activate() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

// stop the timewheel
func (tw *TimeWheel) stop() {
	stopDone := make(chan struct{}, 1)
	tw.stopChannel <- stopDone
	<-stopDone
}

// addTimer adds a new task
// if repeat == true, the timer will be reseted immediately after expired
func (tw *TimeWheel) addTimer(delay time.Duration, callback CallBack, repeat bool) *TimerData {
	if delay <= 0 {
		return nil
	}
	// delay less than interval
	if delay < tw.interval {
		delay = tw.interval
	}
	key := shortuuid.New()
	task := Task{delay: delay, key: key, callback: callback, repeat: repeat}
	tw.addTaskChannel <- &task
	timerData := &TimerData{
		task:      &task,
		Key:       key,
		timewheel: tw,
	}
	return timerData
}

func (tw *TimeWheel) removeTimer(key string) {
	if key == "" {
		return
	}
	tw.removeTaskChannel <- key
}

func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(task)
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case stopFinished := <-tw.stopChannel:
			tw.ticker.Stop()
			stopFinished <- struct{}{}
			return
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.curPos]
	tw.scanAndExcTask(l)
	tw.curPos = (tw.curPos + 1) % tw.slotNum
}

func (tw *TimeWheel) scanAndExcTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}

		go task.callback()
		next := e.Next()
		l.Remove(e)
		if task.repeat {
			tw.addTask(task)
		} else {
			delete(tw.timer, task.key)
		}
		e = next
	}
}

func (tw *TimeWheel) addTask(task *Task) {
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle

	tw.slots[pos].PushBack(task)

	tw.timer[task.key] = Record{
		pos:  pos,
		task: task,
	}
}

func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int64, circle int64) {
	delayNanoseconds := d.Nanoseconds()
	intervalNanoseconds := tw.interval.Nanoseconds()
	circle = delayNanoseconds / intervalNanoseconds / tw.slotNum
	pos = (tw.curPos + delayNanoseconds/intervalNanoseconds) % tw.slotNum

	return
}

func (tw *TimeWheel) removeTask(key string) {
	// 获取定时器所在的槽
	record, ok := tw.timer[key]
	if !ok {
		return
	}
	// 获取槽指向的链表
	l := tw.slots[record.pos]
	for e := l.Front(); e != nil; e = e.Next() {
		task := e.Value.(*Task)
		if task.key == key {
			delete(tw.timer, task.key)
			l.Remove(e)
			break
		}
	}
}

func (td *TimerData) Restart() {
	key := td.task.key
	td.timewheel.removeTimer(key)
	td.timewheel.addTaskChannel <- td.task
}

func (td *TimerData) Remove() {
	key := td.task.key
	td.timewheel.removeTimer(key)
}

func AddTimer(delay time.Duration, callback CallBack) *TimerData {
	return tw.addTimer(delay, callback, false)
}

func AddRepeatTimer(delay time.Duration, callback CallBack) *TimerData {
	return tw.addTimer(delay, callback, true)
}

func RemoveTimer(key string) {
	tw.removeTimer(key)
}

func Stop() {
	tw.stop()
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
