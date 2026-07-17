package domain

import (
	"datasource"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyUserPass(t *testing.T) {
	type args struct {
		username string
		password string
		db       datasource.Database
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestVerifyUserPass",
			args: args{
				username: "test",
				password: "test",
				db:       datasource.NewMapDatabase(),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, _ := bcrypt.GenerateFromPassword([]byte(tt.args.password), bcrypt.DefaultCost)
			newUser := datasource.User{
				Username:     tt.args.username,
				UUID:         "1",
				HashPassword: pass,
			}
			tt.args.db.AddUser(newUser)
			if got := VerifyUserPass(tt.args.username, tt.args.password, tt.args.db); got != tt.want {
				t.Errorf("VerifyUserPass() = %v, want %v", got, tt.want)
			}
		})
	}
}
