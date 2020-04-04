package util

func GetIndex(page, limit int64) int64 {
	// Index-from-limit algorithm formula
	// f(x)= w-x
	// w (omega) = x*n
	// where x = limit and n = page
	return limit*page - limit
}
