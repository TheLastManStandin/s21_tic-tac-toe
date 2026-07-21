package domain

import "datasource"

func gameFromDatasourceToDomain(datasourceGameInstance datasource.GameInstance, UUID string) GameInstance {
	return GameInstance{
		UUID:  UUID,
		Table: datasourceGameInstance.Table,
	}
}

func (g GameInstance) toDatasource() datasource.GameInstance {
	return datasource.GameInstance{
		Table: g.Table,
	}
}

func (g GameInstance) toMinimax() minimaxGame {
	return minimaxGame{
		table: g.Table,
	}
}

func (g minimaxGame) toDomain(UUID string, state datasource.GameState) GameInstance {
	return GameInstance{
		Table: g.table,
		State: state,
		UUID:  UUID,
	}
}
