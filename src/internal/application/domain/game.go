package domain

type Game struct {
	UUID  string
	Table Table
}

type Table [3][3]int8
