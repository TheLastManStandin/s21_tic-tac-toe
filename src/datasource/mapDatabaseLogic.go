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

func (gd *MapDatabase) GetUserByName(username string) (User, error) {
	user, ok := gd.Users.Load(username)
	if !ok {
		return User{}, errors.New("user not found")
	}
	return user.(User), nil
}

func (gd *MapDatabase) AddUser(user User) error {
	if _, exists := gd.Users.Load(user.Username); exists {
		return errors.New("user already exists")
	}
	gd.Users.Store(user.Username, user)
	return nil
}
