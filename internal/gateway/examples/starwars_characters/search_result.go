package starwars_characters

// union SearchResult = Human | Droid
type searchResult struct {
	result interface{}
}

func (r *searchResult) ToHuman() (human, bool) {
	res, ok := r.result.(human)
	return res, ok
}

func (r *searchResult) ToDroid() (droid, bool) {
	res, ok := r.result.(droid)
	return res, ok
}
