package enums

type RequestIdMode string

const (
	AcceptFromHeader RequestIdMode = "AcceptFromHeader"
	AlwaysGenerate   RequestIdMode = "AlwaysGenerate"
)

func (r RequestIdMode) String() string {
	return string(r)
}
