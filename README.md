# lambda-go-build

## About

In a typical Golang AWS Lambda solution there are multiple `main` functions that need to be built separately. 
This tool walks through your codebase and detects all `main` functions and builds them in parallel. 

## Installation

```shell
go get github.com/gtourkas/lambda-go-build

cd $GOPATH/src/github.com/gtourkas/lambda-go-build 

go install
```

## Usage

An invocation requires the base path of your lambdas source (the ``-s`` argument), and the destination (``-d`` argument) for the output of the build process. 
The directory structure of your lambdas, starting from the lambda source base path. 

```shell
lambda-go-build -s ~/myrepo/src/lambdas -d ~/myrepo/build/lambdas
```

One can set the number of parallel builds through the ``--cb`` argument. 
