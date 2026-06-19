package domain

import "datasource"

type Game struct {
	UUID  string
	Table Table
}

type Table [3][3]int8

//func (g Game) toDatabase() datasource.Game {
//	dbGame := datasource.Game{
//		UUID:  g.UUID,
//		Table: datasource.Table(g.table),
//	}
//
//	return dbGame
//}
