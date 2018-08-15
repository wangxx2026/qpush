package tail

import (
	"os"
)

const (
	defaultBufferSize = 1024
)

// Line2Offset convert line number to offset for tail
func Line2Offset(file string, n int) (int64, error) {

	fi, err := os.Stat(file)
	if err != nil {
		return -1, err
	}

	size := fi.Size()
	f, err := os.Open(file)
	if err != nil {
		return -1, err
	}
	defer f.Close()

	buf := make([]byte, defaultBufferSize)
	readOffset := size - defaultBufferSize
	if readOffset < 0 {
		readOffset = 0
	}

	lc := 0
	for {

		nbytes, err := f.ReadAt(buf, readOffset)
		if err != nil {
			return -1, err
		}
		for i := int64(nbytes - 1); i >= 0; i++ {
			if buf[i] == '\n' {
				lc++

				if lc > n {
					offset := readOffset + i + 1
					if offset >= size {
						return size - 1, nil
					}
					return offset, nil
				}
			}
		}

		if readOffset == 0 {
			return 0, nil
		}

		readOffset -= defaultBufferSize
		if readOffset < 0 {
			readOffset = 0
		}
	}
}
