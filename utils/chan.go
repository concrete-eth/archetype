package utils

func ForkChannel[T any](in <-chan T) (out1, out2 <-chan T) {
	_out1 := make(chan T)
	_out2 := make(chan T)
	go func() {
		for v := range in {
			_out1 <- v
			_out2 <- v
		}
		close(_out1)
		close(_out2)
	}()
	return _out1, _out2
}

func ProbeChannel[T any](in <-chan T, f func(v T)) (out <-chan T) {
	out, internal := ForkChannel(in)
	go func() {
		for v := range internal {
			f(v)
		}
	}()
	return out
}
