package interpreter

type Callable interface {
	arity() int
	call(env Environment, args []any) any
}
