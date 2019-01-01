package imagemonkeydb

type InvalidLabelIdError struct {
	Description string
}

func (e *InvalidLabelIdError) Error() string {
	return e.Description
}