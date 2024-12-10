package enums

type EnableDisable int

const (
	Disabled EnableDisable = iota
	Enabled
)

func (o EnableDisable) String() string {
	switch o {
	case Disabled:
		return "disabled"
	case Enabled:
		return "enabled"
	default:
		return "disabled"
	}
}

type AllowOrNot int

const (
	Allow AllowOrNot = iota
	Disallow
)

func (u AllowOrNot) String() string {
	switch u {
	case Allow:
		return "allow"
	case Disallow:
		return "disallow"
	default:
		return "disallow"
	}
}
