package service

type BoxConfigs struct {
	Secret string
	Web    *Web
	Grpc   *Grpc
}
