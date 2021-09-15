package config

type Log struct {
	Level string `json:"level" validate:"required,oneof=debug info warn error"`
}

type Listener struct {
	Address  string `json:"address" validate:"required,hostname_port"`
	Protocol string `json:"protocol" validate:"required,oneof=http grpc"`
}

type Admin struct {
	Listeners []Listener `json:"listeners" validate:"dive"`
}

type Config struct {
	Log   Log   `json:"log" validate:"dive"`
	Admin Admin `json:"admin" validate:"dive"`
}
