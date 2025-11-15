package team

import (
	input "avito-test-pr-service/internal/domain/ports/input"
	"avito-test-pr-service/internal/infrastructure/logger"
)

type TeamHandler struct {
	teamService input.TeamInputPort
	userService input.UserInputPort
	log         *logger.Logger
}

func NewTeamHandler(teamSvc input.TeamInputPort, userSvc input.UserInputPort, log *logger.Logger) *TeamHandler {
	return &TeamHandler{teamService: teamSvc, userService: userSvc, log: log}
}
