package controller

import "booknest/internal/middleware"

var jwtConfig middleware.JWTConfig

func SetJWTConfig(cfg middleware.JWTConfig) {
	jwtConfig = cfg
}

func getJWTConfig() middleware.JWTConfig {
	return jwtConfig
}
