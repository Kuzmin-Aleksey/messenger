package config

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetConfig(t *testing.T) {
	cfg, err := GetConfig("D:\\Users\\akuzm\\GolandProjects\\messenger\\config\\config.yaml")
	require.NoError(t, err)

	fmt.Printf("%#+v\n", cfg.HttpServer)
	fmt.Printf("%#+v\n", cfg.AuthService)
	fmt.Printf("%#+v\n", cfg.Redis)
	fmt.Printf("%#+v\n", cfg.MySQL)
}
