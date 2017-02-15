package deferutil

// Preventable calls a function unless it's Prevent() method has been called.
type Preventable struct {
	isPrevented bool
}

// Do calls fn if Cancel() has never been called.
func (p *Preventable) Do(fn func()) {
	if !p.isPrevented && fn != nil {
		fn()
	}
}

// Prevent prevents future calls to Do() from doing anything
func (p *Preventable) Prevent() {
	p.isPrevented = true
}
