package hunspell

type FSTOutput interface {
	Common(b FSTOutput) FSTOutput
	Subtract(b FSTOutput) FSTOutput
	Prepend(b FSTOutput) FSTOutput
	Append(b FSTOutput) FSTOutput
	// TODO: Hash() uint
}

// TODO: RuneOutput
// TODO: ByteOutput
// TODO: IntOutput

// TODO: optimize Int32Output
type Int32Output []int32

func (a Int32Output) Common(bb FSTOutput) FSTOutput {
	b, ok := bb.(Int32Output)

	if !ok {
		return nil
	}

	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}

	if i < 1 {
		return nil
	}

	out := make(Int32Output, i)
	copy(out, a[:i])

	return out
}

func (a Int32Output) Subtract(bb FSTOutput) FSTOutput {
	if bb == nil {
		out := make(Int32Output, len(a))
		copy(out, a)
		return out
	}

	b, ok := bb.(Int32Output)

	if !ok {
		return nil
	}

	l := len(a) - len(b)

	if l < 1 {
		return nil
	}

	out := make(Int32Output, l)
	copy(out, a[len(b):])

	return out
}

func (a Int32Output) Prepend(bb FSTOutput) FSTOutput {
	b, ok := bb.(Int32Output)

	if !ok {
		return nil
	}

	l := len(b) + len(a)

	if l < 1 {
		return nil
	}

	out := make(Int32Output, 0, l)
	out = append(out, b...)
	out = append(out, a...)

	return out
}

func (a Int32Output) Append(bb FSTOutput) FSTOutput {
	b, ok := bb.(Int32Output)

	if !ok {
		return nil
	}

	l := len(a) + len(b)

	if l < 1 {
		return nil
	}

	out := make(Int32Output, 0, l)
	out = append(out, a...)
	out = append(out, b...)

	return out
}
