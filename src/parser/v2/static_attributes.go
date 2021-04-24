package imagemonkeyquerylang

type UiView int32

const (
    LabelView  UiView = 0
	AnnotationView UiView = 1
)


func GetStaticQueryAttributes(view UiView) []string {
	if view == LabelView {
		return []string{"image.width", "image.height", "annotation.coverage", "image.unlabeled='true'",
							"image.unlabeled='false'", "image.num_labels", "image.num_open_annotation_tasks"}
	}
	if view == AnnotationView {
		return []string{"image.width", "image.height", "annotation.coverage", "image.unlabeled='true'",
							"image.unlabeled='false'", "image.num_labels", "image.num_open_annotation_tasks", "image.num_annotations"}
	}
	return []string{}
}
