package web

// Configuration of  the webservice
type Configuration struct {
	Domain string `env:"DOMAIN"`
	Port   string `env:"PORT"`
}
