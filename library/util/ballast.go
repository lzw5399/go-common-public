package util

const (
    B  = 1
    KB = 1024 * B
    MB = 1024 * KB
    GB = 1024 * MB
)

// Ballast 使用ballast大对象来减少GC的频次
type Ballast struct {
    ballast []byte
}

func NewBallast(sizeMB int) *Ballast {
    size := sizeMB * MB

    var b Ballast
    const maxSize = 2 * GB
    if size > maxSize {
        size = maxSize
    }
    b.ballast = make([]byte, size)
    return &b
}
