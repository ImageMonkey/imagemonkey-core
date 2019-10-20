package imagemonkeydb

type InvalidLabelIdError struct {
	Description string
}

func (e *InvalidLabelIdError) Error() string {
	return e.Description
}

type InvalidTrendingLabelError struct {
	Description string
}

func (e *InvalidTrendingLabelError) Error() string {
	return e.Description
}

type AuthenticationRequiredError struct {
	Description string
}

func (e *AuthenticationRequiredError) Error() string {
	return e.Description
}
