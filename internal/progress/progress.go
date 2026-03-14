package progress

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Bar struct {
	bar      *mpb.Bar
	progress *mpb.Progress
	Writer   io.Writer

	files   atomic.Int64
	skipped atomic.Int64
	matches atomic.Int64
}

type Options struct {
	Workers     int
	Width       int
	RefreshRate time.Duration
}

func New(opts Options) *Bar {
	if opts.Width == 0 {
		opts.Width = 80
	}
	if opts.RefreshRate == 0 {
		opts.RefreshRate = 100 * time.Millisecond
	}

	p := mpb.New(
		mpb.WithWidth(opts.Width),
		mpb.WithRefreshRate(opts.RefreshRate),
	)

	b := &Bar{progress: p, Writer: p}

	b.bar = p.New(
		-1,
		mpb.SpinnerStyle("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"),
		mpb.PrependDecorators(
			decor.Name("SCANNING "),
			decor.Any(func(_ decor.Statistics) string {
				return fmt.Sprintf("%d files  ", b.files.Load())
			}),
			decor.Elapsed(decor.ET_STYLE_GO),
			decor.Name(fmt.Sprintf("  %d workers", opts.Workers)),
			decor.Any(func(_ decor.Statistics) string {
				skipped := b.skipped.Load()
				matches := b.matches.Load()
				s := ""
				if skipped > 0 {
					s += fmt.Sprintf("  skip %d", skipped)
				}
				if matches > 0 {
					s += fmt.Sprintf("  match %d", matches)
				}
				return s + "  "
			}),
		),
		mpb.AppendDecorators(
			decor.Name(" "),
		),
		mpb.BarFillerOnComplete("✓ done"),
		mpb.BarRemoveOnComplete(),
	)

	return b
}

func (b *Bar) IncrementFile() int64 {
	n := b.files.Add(1)
	b.bar.SetCurrent(n)
	return n
}

func (b *Bar) IncrementSkip() {
	b.skipped.Add(1)
}

func (b *Bar) IncrementMatch() {
	b.matches.Add(1)
}

func (b *Bar) Files() int64   { return b.files.Load() }
func (b *Bar) Skipped() int64 { return b.skipped.Load() }
func (b *Bar) Matches() int64 { return b.matches.Load() }

func (b *Bar) Done() {
	b.bar.SetTotal(b.files.Load(), true)
	time.Sleep(200 * time.Millisecond)
	b.progress.Wait()
	b.Writer = os.Stdout
}

func NewDownloadBar(r io.Reader, total int64, name string) (io.ReadCloser, func()) {
	p := mpb.New(
		mpb.WithWidth(50),
		mpb.WithRefreshRate(100*time.Millisecond),
	)

	bar := p.New(
		total,
		mpb.BarStyle().Lbound("  [").Filler("█").Tip("█").Padding("░").Rbound("]"),
		mpb.PrependDecorators(
			decor.Name(name+"  "),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("  %6.1f / %6.1f  "),
			decor.Percentage(decor.WC{W: 6}),
			decor.Name("  "),
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 30),
		),
		mpb.BarFillerOnComplete("✓ downloaded"),
	)

	proxy := bar.ProxyReader(r)
	wait := func() {
		time.Sleep(200 * time.Millisecond)
		p.Wait()
	}

	return proxy, wait
}
