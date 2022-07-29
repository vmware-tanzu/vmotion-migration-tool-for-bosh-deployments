package vcenter

import (
	"fmt"
	"sync"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi/vim25/progress"
)

type ProgressLogger struct {
	taskReportsCh   chan taskReport
	startedCh       chan struct{}
	openSinksOnce   sync.Once
	openSinksWG     sync.WaitGroup
	updatableStdout *log.UpdatableStdout
}

func NewProgressLogger(updatableStdout *log.UpdatableStdout) *ProgressLogger {
	p := &ProgressLogger{
		taskReportsCh:   make(chan taskReport),
		startedCh:       make(chan struct{}),
		updatableStdout: updatableStdout,
	}
	go p.checkAllSinksComplete()
	go p.updateProgress()
	return p
}

func (p *ProgressLogger) updateProgress() {
	for report := range p.taskReportsCh {
		s := fmt.Sprintf("%s - %.0f%%", report.TaskName, report.Percent)
		if report.Error != nil {
			s = fmt.Sprintf("%s - %s", report.TaskName, report.Error)
		}
		p.updatableStdout.PrintUpdatable(report.TaskName, s)
	}
}

func (p *ProgressLogger) NewProgressSink(taskName string) *ProgressSink {
	p.openSinksWG.Add(1)
	p.openSinksOnce.Do(func() {
		p.startedCh <- struct{}{}
	})
	return &ProgressSink{
		logger:   p,
		taskName: taskName,
	}
}

func (p *ProgressLogger) checkAllSinksComplete() {
	<-p.startedCh
	p.openSinksWG.Wait()
	close(p.taskReportsCh)
}

func (p *ProgressLogger) SinkDone() {
	p.openSinksWG.Done()
}

type ProgressSink struct {
	logger   *ProgressLogger
	taskName string
}

func (p *ProgressSink) Sink() chan<- progress.Report {
	l := log.WithoutContext()
	sinkCh := make(chan progress.Report)
	go func(ps *ProgressSink, sinkCh chan progress.Report) {
		for report := range sinkCh {
			if report.Error() != nil {
				l.Debugf("Received %s progress update: %f: %s", ps.taskName, report.Percentage(), report.Error())
			} else {
				l.Debugf("Received %s progress update: %f", ps.taskName, report.Percentage())
			}

			// throw away non-error 0 progress as vSphere reports 0 _after_ 100 which corrupts completed VMs output
			if report.Percentage() > 0 || report.Error() != nil {
				ps.logger.taskReportsCh <- taskReport{
					Percent:  report.Percentage(),
					Error:    report.Error(),
					Detail:   report.Detail(),
					TaskName: ps.taskName,
				}
			}
		}
		p.logger.SinkDone()
	}(p, sinkCh)
	return sinkCh
}

type taskReport struct {
	Percent  float32
	Error    error
	Detail   string
	TaskName string
}
