package datasource

import (
	"reflect"
	"sync"
	"testing"
)

func TestGamesDatabase_LoadGame(t *testing.T) {
	type fields struct {
		Games sync.Map
	}
	type args struct {
		tableUUID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    DbGame
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gd := &GamesDatabase{
				Games: tt.fields.Games,
			}
			got, err := gd.LoadGame(tt.args.tableUUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadGame() got = %v, want %v", got, tt.want)
			}
		})
	}
}
