package utils

func MustOk(err error) {
	if err != nil {
		panic(err)
	}
}
