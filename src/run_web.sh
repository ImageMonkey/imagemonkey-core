#!/bin/bash

go build -o web web.go web_secrets.go auth.go && ./web
