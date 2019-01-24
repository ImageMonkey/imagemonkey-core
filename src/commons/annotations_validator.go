package commons

import (
	"encoding/json"
	datastructures "../datastructures"
	"bytes"
	"github.com/getsentry/raven-go"
	"errors"
)

type AnnotationsValidator struct {
	annotations []json.RawMessage
	annotationRefinements [][]datastructures.AnnotationRefinementEntry
}

func NewAnnotationsValidator(annotations []json.RawMessage) *AnnotationsValidator {
    return &AnnotationsValidator {
    	annotations: annotations,
    	annotationRefinements: [][]datastructures.AnnotationRefinementEntry{},
    }
}

func (p *AnnotationsValidator) Parse() error {
	for _, r := range p.annotations {
		var obj map[string]interface{}
        err := json.Unmarshal(r, &obj)
        if err != nil {
            return err
        }

        shapeType := obj["type"]
        if shapeType == "rect" {
        	var rectangleAnnotation datastructures.RectangleAnnotation 
        	decoder := json.NewDecoder(bytes.NewReader([]byte(r)))
        	decoder.DisallowUnknownFields() //throw an error in case of an unknown field 
        	err = decoder.Decode(&rectangleAnnotation)
        	if err != nil {
        		raven.CaptureError(err, nil)
        		return err
        	}
        	p.annotationRefinements = append(p.annotationRefinements, rectangleAnnotation.Refinements)
        } else if shapeType == "ellipse" {
        	var ellipsisAnnotation datastructures.EllipsisAnnotation 
			decoder := json.NewDecoder(bytes.NewReader([]byte(r)))
        	decoder.DisallowUnknownFields() //throw an error in case of an unknown field 
        	err = decoder.Decode(&ellipsisAnnotation)
        	if err != nil {
        		raven.CaptureError(err, nil)
        		return err
        	}
        	p.annotationRefinements = append(p.annotationRefinements, ellipsisAnnotation.Refinements)
        } else if shapeType == "polygon" {
        	var polygonAnnotation datastructures.PolygonAnnotation 
			decoder := json.NewDecoder(bytes.NewReader([]byte(r)))
        	decoder.DisallowUnknownFields() //throw an error in case of an unknown field 
        	err = decoder.Decode(&polygonAnnotation)
        	if err != nil {
        		raven.CaptureError(err, nil)
        		return err
        	}
        	p.annotationRefinements = append(p.annotationRefinements, polygonAnnotation.Refinements)
        } else {
        	return errors.New("Invalid shape type")
        }
	}

	if len(p.annotations) == 0 {
		return errors.New("annotations missing")
	}

	return nil
}

func (p *AnnotationsValidator) GetRefinements() [][]datastructures.AnnotationRefinementEntry {
	return p.annotationRefinements
}