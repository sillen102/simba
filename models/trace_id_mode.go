package models

type TraceIDMode string

const (
	AcceptFromHeader TraceIDMode = "AcceptFromHeader"
	AlwaysGenerate   TraceIDMode = "AlwaysGenerate"
)

func (r TraceIDMode) String() string {
	return string(r)
}
