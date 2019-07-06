#!/bin/bash

go build -o api api.go auth.go label_graph.go && ./api
