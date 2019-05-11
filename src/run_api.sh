#!/bin/bash

go build -o api api.go api_secrets.go auth.go label_graph.go && ./api
