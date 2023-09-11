package custom

import "sync"

type Buffer []byte

// sync.Pool は個別に保存したり、取り出したりできる一時的なオブジェクトの集合
// Pool に保存されたアイテムは任意のタイミングでなんの通知もなしに自動的に削除されるかもしれない
// Pool は goroutine safe になっている（複数goroutineが同時使用しても安全）
var bufPool = sync.Pool{
	New: func() any {
		// slogのcommonHandler参考にあらかじめある程度容量を確保しておく
		// なぜこの値が妥当なのかはわからない...
		b := make(Buffer, 0, 1024)
		return &b
	},
}

func NewBuffer() *Buffer {
	return bufPool.Get().(*Buffer)
}

func (b *Buffer) Free() {
	// ピーク時の割り当てを減らすには、小さいバッファだけをプールに戻す。
	// slogのcommonHandler参考にこの処理を入れている
	// 空にするために実体を取得しているから巨大なバッファを処理するのは避けたいってこと？
	const maxBufferSize = 16 << 10
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bufPool.Put(b)
	}
}

func (b *Buffer) Write(p []byte) {
	*b = append(*b, p...)
}

func (b *Buffer) WriteByte(c byte) {
	*b = append(*b, c)
}

func (b *Buffer) WriteString(s string) {
	*b = append(*b, s...)
}
