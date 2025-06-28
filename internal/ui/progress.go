package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

var (
	gray  = color.New(color.FgHiBlack)
	green = color.New(color.FgGreen)
	white = color.New(color.FgWhite, color.Bold)
	cyan  = color.New(color.FgCyan)
)

type ProgressDisplay struct {
	stopSpinner chan bool
	isSpinning  bool
}

func NewProgressDisplay() *ProgressDisplay {
	return &ProgressDisplay{
		stopSpinner: make(chan bool),
	}
}

func (p *ProgressDisplay) StartSection(emoji, title string) {
	fmt.Printf("┌─ %s %s\n", emoji, white.Sprint(title))
	fmt.Println("│")
}

func (p *ProgressDisplay) ShowCacheStatus(found int, lastSync string) {
	fmt.Printf("├─ %s %s\n", "📋", gray.Sprint("Checking cache..."))
	if found > 0 {
		fmt.Printf("│  %s Found %s cached %s\n",
			green.Sprint("✓"),
			white.Sprintf("%d PRs", found),
			gray.Sprintf("(last sync: %s)", lastSync))
	} else {
		fmt.Printf("│  %s No cache found\n", gray.Sprint("○"))
	}
	fmt.Println("│")
}

func (p *ProgressDisplay) StartFetching() {
	fmt.Printf("├─ %s %s\n", "📡", gray.Sprint("Fetching from GitHub..."))
}

func (p *ProgressDisplay) ShowProgress(label string, count int, unit string) {
	p.stopCurrentSpinner()
	p.startSpinner(label, count, unit)
}

func (p *ProgressDisplay) StopProgress() {
	p.stopCurrentSpinner()
}

func (p *ProgressDisplay) ShowSuccess(count int, filename, size string) {
	fmt.Printf("│\n")
	fmt.Printf("└─ %s Exported %s → %s %s\n",
		green.Sprint("✅"),
		white.Sprintf("%d PRs", count),
		cyan.Sprint(filename),
		gray.Sprintf("(%s)", size))
	fmt.Println()
	fmt.Printf("%s Analysis ready! Try: %s\n",
		"🎉",
		gray.Sprintf("duckdb -c \"SELECT * FROM '%s'\"", filename))
}

func (p *ProgressDisplay) ShowError(err error) {
	p.stopCurrentSpinner()
	fmt.Printf("│\n")
	fmt.Printf("└─ %s Error: %s\n", "❌", err.Error())
}

func (p *ProgressDisplay) startSpinner(label string, count int, unit string) {
	p.stopSpinner = make(chan bool)
	p.isSpinning = true

	go func() {
		spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-p.stopSpinner:
				// Clear the line and show final result
				fmt.Printf("\r│  ├─ %s %s %s %s\n",
					gray.Sprint(label+"................"),
					green.Sprint("✓"),
					white.Sprintf("%d", count),
					gray.Sprint(unit))
				return
			case <-ticker.C:
				fmt.Printf("\r│  ├─ %s %s %s %s",
					gray.Sprint(label+"................"),
					cyan.Sprint(spinners[i]),
					white.Sprintf("%d", count),
					gray.Sprint(unit))
				i = (i + 1) % len(spinners)
			}
		}
	}()
}

func (p *ProgressDisplay) stopCurrentSpinner() {
	if p.isSpinning {
		p.stopSpinner <- true
		p.isSpinning = false
		time.Sleep(50 * time.Millisecond) // Give time for the goroutine to finish
	}
}
