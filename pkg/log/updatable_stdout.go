package log

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

// UpdatableStdout is used to print and update text to stdout. Use this instead of fmt.Print to support
// updating already output text like progress percentage.
type UpdatableStdout struct {
	priorLinesPrinted int
	sync.RWMutex
	ids   []string
	lines map[string]string
}

// NewUpdatableStdout creates a new UpdatableStdout instance that can be used to print and update text to stdout
func NewUpdatableStdout() *UpdatableStdout {
	return &UpdatableStdout{
		lines: make(map[string]string),
	}
}

// Println outputs an empty line
func (p *UpdatableStdout) Println() {
	p.Print("")
}

// Print outputs the specified string as-is on the current line
func (p *UpdatableStdout) Print(s string) {
	id := uuid.New()
	p.PrintUpdatable(id.String(), s)
}

// Printf formats and outputs the format string on the current line
func (p *UpdatableStdout) Printf(format string, a ...interface{}) {
	p.Print(fmt.Sprintf(format, a...))
}

// PrintUpdatablef formats and outputs the format string on the current line, or updates it in place
func (p *UpdatableStdout) PrintUpdatablef(id, format string, a ...interface{}) {
	p.PrintUpdatable(id, fmt.Sprintf(format, a...))
}

// PrintUpdatable outputs the specified string on the current line, or updates it in place
func (p *UpdatableStdout) PrintUpdatable(id, s string) {
	p.Lock()
	defer p.Unlock()

	var exists bool
	if _, exists = p.lines[id]; !exists {
		p.ids = append(p.ids, id)
	}
	p.lines[id] = s
	p.flush()
}

func (p *UpdatableStdout) flush() {
	if p.priorLinesPrinted > 0 {
		fmt.Printf("\033[%dF", p.priorLinesPrinted)
	}

	p.priorLinesPrinted = 0
	for _, id := range p.ids {
		p.priorLinesPrinted++
		fmt.Println(p.lines[id])
	}
}
