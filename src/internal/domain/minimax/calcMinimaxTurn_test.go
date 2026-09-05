package minimax

import (
	"testing"
)

func Test_minimax1(t *testing.T) {
	type args struct {
		game *minimaxGame
		turn int
	}
	tests := []struct {
		name string
		args args
		want minimaxGame
	}{
		{
			name: "case 1",
			args: args{
				game: &minimaxGame{
					table: [3][3]int{
						{0, 0, 2},
						{2, 1, 1},
						{1, 1, 2},
					},
				},
				turn: 2,
			},
			want: minimaxGame{
				table: [3][3]int{
					{0, 2, 2},
					{2, 1, 1},
					{1, 1, 2},
				},
			},
		},
		{
			name: "case 2",
			args: args{
				game: &minimaxGame{
					table: [3][3]int{
						{0, 0, 0},
						{0, 1, 0},
						{0, 0, 0},
					},
				},
				turn: 2,
			},
			want: minimaxGame{
				table: [3][3]int{
					{2, 0, 0},
					{0, 1, 0},
					{0, 0, 0},
				},
			},
		},
		{
			name: "case 3",
			args: args{
				game: &minimaxGame{
					table: [3][3]int{
						{0, 0, 0},
						{0, 2, 1},
						{2, 1, 1},
					},
				},
				turn: 2,
			},
			want: minimaxGame{
				table: [3][3]int{
					{0, 0, 2},
					{0, 2, 1},
					{2, 1, 1},
				},
			},
		},
		{
			name: "case 4",
			args: args{
				game: &minimaxGame{
					table: [3][3]int{
						{0, 0, 1},
						{1, 2, 2},
						{1, 0, 0},
					},
				},
				turn: 2,
			},
			want: minimaxGame{
				table: [3][3]int{
					{2, 0, 1},
					{1, 2, 2},
					{1, 0, 0},
				},
			},
		},
		//{
		//	name: "case 5",
		//	args: args{
		//		game: &minimaxGame{
		//			table: [3][3]int{
		//				{0, 1, 0},
		//				{0, 2, 0},
		//				{2, 1, 1},
		//			},
		//		},
		//		turn: 2,
		//	},
		//	want: minimaxGame{
		//		table: [3][3]int{
		//			{0, 1, 2},
		//			{0, 2, 0},
		//			{2, 1, 1},
		//		},
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minimax(tt.args.game, tt.args.turn)
			if tt.args.game.table != tt.want.table {
				t.Errorf("minimax() = %v, want %v", *tt.args.game, tt.want)
			}
		})
	}
}
