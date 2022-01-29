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

type DuplicateImageCollectionError struct {
	Description string
}

func (e *DuplicateImageCollectionError) Error() string {
	return e.Description
}

type InvalidImageCollectionInputError struct {
	Description string
}

func (e *InvalidImageCollectionInputError) Error() string {
	return e.Description
}

func (e *NotFoundError) Error() string {
	return e.Description
}

type NotFoundError struct {
	Description string
}
