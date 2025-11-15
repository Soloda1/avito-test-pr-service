package pr

import (
	input "avito-test-pr-service/internal/domain/ports/input"
	ports "avito-test-pr-service/internal/domain/ports/output"
)

type PRHandler struct {
	prService input.PRInputPort
	log       ports.Logger
}

func NewPRHandler(s input.PRInputPort, log ports.Logger) *PRHandler {
	return &PRHandler{prService: s, log: log}
}
