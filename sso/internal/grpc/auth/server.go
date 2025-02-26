package auth

import ssov1 "contracts/proto/sso"

func NewServerAPI() ssov1.AuthV1Server {
	return &serverAPI{}
}

type serverAPI struct {
}
