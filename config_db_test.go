package tgo

import (
	"testing"
)

func Test_ConfigDbInit(t *testing.T) {
	configDbInit()

	t.Errorf("config db is %v", dbConfig.Mysql.Pool.PoolIdleTimeout)
}
