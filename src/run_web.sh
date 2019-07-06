#!/bin/bash

go build -o web web.go auth.go && ./web
