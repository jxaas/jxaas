package scheduler

import (
	"time"

	"github.com/justinsb/gova/log"
)

type ScheduledTask struct {
	scheduler *Scheduler
	task      Runnable
	interval  time.Duration
}

type Scheduler struct {
}

func NewScheduler() *Scheduler {
	self := &Scheduler{}
	return self
}

func (self *Scheduler) AddTask(task Runnable, interval time.Duration) *ScheduledTask {
	scheduledTask := &ScheduledTask{}
	scheduledTask.task = task
	scheduledTask.scheduler = self
	scheduledTask.interval = interval

	go scheduledTask.run()

	return scheduledTask
}

func (self *ScheduledTask) run() {
	for {
		time.Sleep(self.interval)
		err := self.task.Run()
		if err != nil {
			log.Warn("Error running task %v", self.task, err)
		}
	}
}
