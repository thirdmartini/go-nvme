package buffers

type Buffer []byte

type Buffers struct {
	c chan Buffer
}

func New(count int, size int) *Buffers {
	b := &Buffers{
		c: make(chan Buffer, count),
	}

	for i := 0; i < count; i++ {
		buf := make(Buffer, size, size)
		b.c <- buf
	}

	return b
}

func (b *Buffers) Put(buf Buffer) {
	b.c <- buf
}

func (b *Buffers) Get() Buffer {
	return <-b.c
}
