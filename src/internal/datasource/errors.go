package datasource

type DatabaseGameNotFount struct {
}

func (DatabaseGameNotFount) Error() string {
	return "game not fount"
}
