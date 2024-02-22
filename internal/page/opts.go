package page

type HandleOpt byte

const (
	None HandleOpt = 1 << iota
	RequireSession
	CanGoBack
	RequireAdminPermission
)
