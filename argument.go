package pg

type Argument struct {
	key   string
	value any
}

func Arg(key string, value any) *Argument {
	return &Argument{
		key,
		value,
	}
}
