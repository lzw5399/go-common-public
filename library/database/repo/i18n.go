package repo

func IgnoreI18n(skipI18n bool) RepoOptionFunc {
	return func(option *RepoOption) {
		option.ignoreI18n = skipI18n
	}
}
