package log

type BufPool struct {
	pool   [][]byte
	cursor int
	len    int
}

func NewBufPool(size, len int) BufPool {
	pool := make([][]byte, 0, len)
	for range len {
		buf := make([]byte, 0, size)
		pool = append(pool, buf)
	}
	return BufPool{pool, -1, len}
}

func (p *BufPool) Next() []byte {
	p.cursor += 1
	if p.cursor >= p.len {
		p.cursor = 0
	}

	return p.pool[p.cursor][:0]
}
