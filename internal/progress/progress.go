package progress

import (
	"fmt"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Bar struct {
	bar      *mpb.Bar
	progress *mpb.Progress
}

type Options struct {
	Workers     int
	Width       int
	RefreshRate time.Duration
}

func New(opts Options) *Bar {
	if opts.Width == 0 {
		opts.Width = 50
	}
	if opts.RefreshRate == 0 {
		opts.RefreshRate = 100 * time.Millisecond
	}

	p := mpb.New(
		mpb.WithWidth(opts.Width),
		mpb.WithRefreshRate(opts.RefreshRate),
	)

	bar := p.New(
		-1,
		mpb.SpinnerStyle("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"),
		mpb.PrependDecorators(
			decor.Name("SCANNING "),
			decor.CurrentNoUnit("%d files  "),
			decor.Elapsed(decor.ET_STYLE_GO),
			decor.Name(fmt.Sprintf("  %d workers  ", opts.Workers)),
		),
		mpb.AppendDecorators(
			decor.Name(" "),
		),
		mpb.BarFillerOnComplete("✓ done"),
		mpb.BarRemoveOnComplete(),
	)

	return &Bar{bar: bar, progress: p}
}

func (b *Bar) Increment(n int64) {
	b.bar.SetCurrent(n)
}

// Done đánh dấu bar hoàn thành và đợi render frame cuối
func (b *Bar) Done(total int64) {
	b.bar.SetTotal(total, true)
	time.Sleep(200 * time.Millisecond)
	b.progress.Wait()
}
