package lua

type Script interface {
	Name() string
	Function() string
}
