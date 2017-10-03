package main

import (
	//tf_core_framework "tensorflow/core/framework"
	//pb "tensorflow_serving/apis"

	framework "tensorflow/core/framework"
	pb "tensorflow_serving"

	google_protobuf "github.com/golang/protobuf/ptypes/wrappers"
	//tf "github.com/tensorflow/tensorflow/tensorflow/go"

	"google.golang.org/grpc"
)


func predict(imageBytes bytes[]){
	/*tensor, err := tf.NewTensor(string(imageBytes))
	if err != nil {
		log.Debug("Cannot read image file")
	}

	tensorString, ok := tensor.Value().(string)
	if !ok {
		log.Debug("Cannot type assert tensor value to string")
	}*/
}