package util

type ILogger interface {
	Print(message, resource string)
	Error(message, resource string)
	Fatal(message, resource string)
}
