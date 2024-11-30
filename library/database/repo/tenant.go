package repo

func IgnoreTenant(ignoreTenant bool) RepoOptionFunc {
	return func(option *RepoOption) {
		option.ignoreTenant = ignoreTenant
	}
}
