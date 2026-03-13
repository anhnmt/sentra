package progress

import (
	"fmt"
	"io"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Bar struct {
	bar      *mpb.Bar
	progress *mpb.Progress
	Writer   io.Writer // zerolog ghi vào đây thay vì os.Stdout
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

	return &Bar{
		bar:      bar,
		progress: p,
		Writer:   p, // mpb.Progress implement io.Writer, tự xử lý clear/redraw
	}
}

func (b *Bar) Increment(n int64) {
	b.bar.SetCurrent(n)
}

func (b *Bar) Done(total int64) {
	b.bar.SetTotal(total, true)
	time.Sleep(200 * time.Millisecond)
	b.progress.Wait()
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
