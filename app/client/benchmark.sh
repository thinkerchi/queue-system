#!/bin/bash

num=$1

for ((i=0;i<num;i++)) {
	go run client.go &
}
