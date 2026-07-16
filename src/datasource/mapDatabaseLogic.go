package datasource

import "errors"

func NewMapDatabase() *MapDatabase {
	db := &MapDatabase{}
	return db
}

func (gd *MapDatabase) SaveGame(game GameInstance, UUID string) error {
	gd.Games.Store(UUID, game)
	return nil
}

func (gd *MapDatabase) LoadGame(UUID string) (GameInstance, error) {
	game, ok := gd.Games.Load(UUID)
	if !ok {
		return GameInstance{}, errors.New("game not found")
	}

	return game.(GameInstance), nil
}

func (gd *MapDatabase) Close() error {
	return nil
}
