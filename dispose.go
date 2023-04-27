package di

// IDisposable 可处置
type IDisposable interface{ OnDispose() error }
