package datasource

type GameRow struct {
	UUID      string
	GameTable [][]int32
	GameState int
}

type UserRow struct {
	UUID     string
	Username string
	Password string
}
