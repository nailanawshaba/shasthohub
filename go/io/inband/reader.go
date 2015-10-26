package inband

import (
	"errors"
	"io"
)

// String-matching in Golang using the Knuth–Morris–Pratt algorithm (KMP)
// Implementation details from: https://github.com/paddie/gokmp
type inbandReader struct {
	err     error
	stopSeq []byte
	prefix  []int
	r       io.Reader
	scratch []byte
	out     []byte // leftover consumed output
	m       int    // KMP parameters
	i       int    // KMP parameters
}

func NewReader(stopSeq []byte, r io.Reader) (io.Reader, error) {
	prefix, err := computePrefix(stopSeq)
	if err != nil {
		return nil, err
	}
	return &inbandReader{
		stopSeq: stopSeq,
		prefix:  prefix,
		r:       r,
		scratch: make([]byte, len(stopSeq)),
	}, nil
}

// returns an array containing indexes of matches
// - error if pattern argument is less than 1 char
func computePrefix(pattern []byte) ([]int, error) {
	// sanity check
	lenp := len(pattern)
	if len_p < 2 {
		if len_p == 0 {
			return nil, errors.New("'pattern' must contain at least one character")
		}
		return []int{-1}, nil
	}
	t := make([]int, lenp)
	t[0], t[1] = -1, 0

	pos, count := 2, 0
	for pos < lenp {

		if pattern[pos-1] == pattern[count] {
			count++
			t[pos] = count
			pos++
		} else {
			if count > 0 {
				count = t[count]
			} else {
				t[pos] = 0
				pos++
			}
		}
	}
	return t, nil
}

func (ibr *inbandReader) read(p []byte) (int, bool, error) {
	iStart := ibr.i
	n, err := ibr.r.Read(p[ibr.i:])

	// In the case of an EOF, be sure to copy out any partial
	// matches that we had in progress
	if err == io.EOF {
		ret := copy(p, ibr.stopSeq[:ibr.i])
		return ret, true, nil
	}

	// All other errors have us bailing out immediately
	if err != nil {
		return 0, false, err
	}

	for ibr.m+ibr.i < n {
		if ibr.stopSeq[ibr.i] == p[ibr.m+ibr.i] {
			if ibr.i == len(ibr.stopSeq)-1 {
				return ibr.m, true, nil
			}
			ibr.i++
		} else {
			ibr.m = ibr.m + ibr.i - ibr.prefix[ibr.i]
			if ibr.prefix[ibr.i] > -1 {
				ibr.i = ibr.prefix[ibr.i]
			} else {
				ibr.i = 0
			}
		}
	}

	// We can output all characters that aren't partial matches.
	// We'll deal with the partial matches on the next read
	ret := ibr.m
	ibr.m = 0
	copy(p[0:iStart], ibr.stopSeq)

	return ret, false, nil
}

// Read fulfills the read contract, reading until the EOF in a stream
// or until the stop sequence is encountered.
func (ibr *inbandReader) Read(p []byte) (int, error) {
	if ibr.err != nil && ibr.err != io.EOF {
		return 0, ibr.err
	}

	// Output the leftovers if there were any.  We can still do this
	// even if we've hit an EOF condition
	if len(ibr.out) > 0 {
		ret := copy(p, ibr.out)
		ibr.out = ibr.out[ret:]
		return ret, nil
	}

	// If we don't have any more data to read out, then say we've hit
	// the EOF
	if ibr.err == io.EOF {
		return 0, ibr.err
	}

	// By default, we can just ready right into the given buffer...
	buf := p

	// BUT! We might have a "short" buffer situation, in which the passed-in
	// buffer is shorter than our stop sequence.
	shortBuf := false
	if len(buf) < len(ibr.stopSeq) {
		buf = ibr.scratch
		shortBuf = true
	}

	// Actually read, updating the KMP state machine as necessary
	ret, eof, err := ibr.read(buf)

	if err != nil {
		ibr.err = err
		return 0, err
	}

	// EOF conditions are dealt with in future iterations if there's still
	// data to be read out of the reader right now
	if eof {
		ibr.err = io.EOF
		if ret == 0 {
			return 0, io.EOF
		}
	}

	// If we didn't need a short buffer, we're OK to return right away
	if !shortBuf {
		return ret, nil
	}

	// In the short buffer situation, copy as much as we can into the
	// given buffer, and save the rest for the next iteration
	ret = copy(p, buf)
	ibr.out = buf[ret:]
	return ret, nil
}
