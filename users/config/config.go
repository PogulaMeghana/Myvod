package config

type Properties struct {
	Port           string `env:"USERS_SRV_PORT" env-default:"51000"`
	Host           string `env:"USERS_HOST" env-default:"localhost"`
	DBUser         string `env:"DB_USER" env-default:""`
	DBPass         string `env:"DB_PASS" env-default:""`
	DBName         string `env:"DB_NAME" env-default:""`
	DBURL          string `env:"DB_URL" env-default:""`
	UserCollection string `env:"USER_COLLECTION" env-default:"users"`
	JwtTokenSecret string `env:"JWT_TOKEN_SECRET" env-default:""`
}
