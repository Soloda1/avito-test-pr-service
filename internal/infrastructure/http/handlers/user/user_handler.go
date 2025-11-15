package user

import (
	input "avito-test-pr-service/internal/domain/ports/input"
	"avito-test-pr-service/internal/infrastructure/logger"
)

type UserHandler struct {
	userService input.UserInputPort
	prService   input.PRInputPort
	log         *logger.Logger
}

func NewUserHandler(userSvc input.UserInputPort, prSvc input.PRInputPort, log *logger.Logger) *UserHandler {
	return &UserHandler{userService: userSvc, prService: prSvc, log: log}
}
