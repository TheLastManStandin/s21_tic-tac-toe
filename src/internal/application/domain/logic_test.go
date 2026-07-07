package domain

import (
	"datasource"
	"testing"
)

func TestService_GameEnded(t *testing.T) {
	type fields struct {
		db *datasource.GameRepository
	}
	type args struct {
		game Game
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "case 1",
			fields: fields{
				db: nil,
			},
			args: args{
				game: Game{
					UUID: "",
					Table: Table{
						{0, 0, 0},
						{0, 0, 0},
						{0, 0, 0},
					},
				},
			},
			want: false,
		},
		{
			name: "case 2",
			fields: fields{
				db: nil,
			},
			args: args{
				game: Game{
					UUID: "",
					Table: Table{
						{1, 0, 0},
						{0, 1, 0},
						{0, 0, 1},
					},
				},
			},
			want: true,
		},
		{
			name: "case 1",
			fields: fields{
				db: nil,
			},
			args: args{
				game: Game{
					UUID: "",
					Table: Table{
						{0, 0, 0},
						{1, 1, 1},
						{0, 0, 0},
					},
				},
			},
			want: true,
		},
		{
			name: "case 1",
			fields: fields{
				db: nil,
			},
			args: args{
				game: Game{
					UUID: "",
					Table: Table{
						{0, 0, 1},
						{0, 0, 1},
						{0, 0, 1},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{
				db: tt.fields.db,
			}
			if got := s.GameEnded(tt.args.game); got != tt.want {
				t.Errorf("GameEnded() = %v, want %v", got, tt.want)
			}
		})
	}
}
