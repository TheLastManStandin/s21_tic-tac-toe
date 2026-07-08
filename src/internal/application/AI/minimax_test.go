package AI

import (
	"reflect"
	"testing"
)

func TestMinimax(t *testing.T) {
	type args struct {
		game MinimaxGame
		turn int
	}
	tests := []struct {
		name string
		args args
		want MinimaxGame
	}{
		{
			name: "case 1",
			args: args{
				game: MinimaxGame{
					Table: [3][3]int8{
						{0, 0, 2},
						{2, 1, 1},
						{1, 1, 2},
					},
				},
				turn: 2,
			},
			want: MinimaxGame{
				Table: [3][3]int8{
					{0, 2, 2},
					{2, 1, 1},
					{1, 1, 2},
				},
				State: 1,
			},
		},
		{
			name: "case 2",
			args: args{
				game: MinimaxGame{
					Table: [3][3]int8{
						{0, 0, 0},
						{0, 1, 0},
						{0, 0, 0},
					},
				},
				turn: 2,
			},
			want: MinimaxGame{
				Table: [3][3]int8{
					{2, 0, 0},
					{0, 1, 0},
					{0, 0, 0},
				},
				State: 1,
			},
		},
		{
			name: "case 3",
			args: args{
				game: MinimaxGame{
					Table: [3][3]int8{
						{0, 0, 0},
						{0, 2, 1},
						{2, 1, 1},
					},
				},
				turn: 2,
			},
			want: MinimaxGame{
				Table: [3][3]int8{
					{0, 0, 2},
					{0, 2, 1},
					{2, 1, 1},
				},
				State: 1,
			},
		},
		{
			name: "case 4",
			args: args{
				game: MinimaxGame{
					Table: [3][3]int8{
						{0, 0, 1},
						{1, 2, 2},
						{1, 0, 0},
					},
				},
				turn: 2,
			},
			want: MinimaxGame{
				Table: [3][3]int8{
					{2, 0, 1},
					{1, 2, 2},
					{1, 0, 0},
				},
				State: 1,
			},
		},
		{
			name: "case 5",
			args: args{
				game: MinimaxGame{
					Table: [3][3]int8{
						{2, 0, 0},
						{0, 1, 0},
						{0, 0, 1},
					},
				},
				turn: 2,
			},
			want: MinimaxGame{
				Table: [3][3]int8{
					{2, 0, 2},
					{0, 1, 0},
					{0, 0, 1},
				},
				State: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Minimax(tt.args.game, tt.args.turn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Minimax() = %v, want %v", got, tt.want)
			}
		})
	}
}
