package runner

type Runner struct {
	options *Options
}

func New(options *Options) (*Runner, error) {
	runner := &Runner{
		options: options,
	}

	return runner, nil
}

func (r *Runner) Run() {

}
