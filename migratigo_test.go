package migratigo

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestConnect(t *testing.T) {
	type args struct {
		host     string
		port     string
		username string
		password string
		name     string
	}
	tests := []struct {
		name    string
		args    args
		want    *sql.DB
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				host:     "localhost",
				port:     "5432",
				username: "postgres",
				password: "postgres",
				name:     "migratico",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Connect(tt.args.host, tt.args.port, tt.args.username, tt.args.password, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Connect() got = %v, want %v", got, tt.want)
			}
		})
	}
}
