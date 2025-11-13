package models

type PRStatus string

const (
	PRStatusOPEN   PRStatus = "OPEN"
	PRStatusMERGED PRStatus = "MERGED"
)
