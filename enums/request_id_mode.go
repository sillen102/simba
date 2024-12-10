package enums

type RequestIdMode int

const (
	AcceptFromHeader RequestIdMode = iota
	AlwaysGenerate
)

func (r RequestIdMode) String() string {
	switch r {
	case AcceptFromHeader:
		return "AcceptFromHeader"
	case AlwaysGenerate:
		return "AlwaysGenerate"
	default:
		return "AcceptFromHeader"
	}
}
