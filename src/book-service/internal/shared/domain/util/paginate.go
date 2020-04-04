package util

func GetIndex(page, limit int64) int64 {
	// Index-from-limit algorithm formula
	// f(x)= w-x
	// w (omega) = xy
	// where x = limit and y = page
	return limit*page - limit
}
