package ports

//go:generate mockery --name Logger --dir . --output ../../../../mocks --outpkg mocks --with-expecter --filename Logger.go

type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}
