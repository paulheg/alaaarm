package web

// Configuration of  the webservice
type Configuration struct {
	Domain        string `env:"DOMAIN"`
	Port          string `env:"PORT"`
	ViewDirectory string `env:"VIEW_DIR"`
}
