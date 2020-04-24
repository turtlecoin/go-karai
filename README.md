# [go-karai]

## Karai is a highly scalable transaction graph for distributed applications

![karai_github_banner](https://user-images.githubusercontent.com/34389545/80034381-f6a14d00-84b3-11ea-857a-638322dac890.png)

[![Discord](https://img.shields.io/discord/388915017187328002?label=Join%20Discord)](http://chat.turtlecoin.lol) [![GitHub issues](https://img.shields.io/github/issues/turtlecoin/go-karai?label=Issues)](https://github.com/turtlecoin/go-karai/issues) ![GitHub stars](https://img.shields.io/github/stars/turtlecoin/go-karai?label=Github%20Stars)

## Usage

`./go-karai`

This will launch `go-karai`

Type `menu` to view a list of functions. Functions that are darkened are disabled.

## Dependencies

-   Golang 1.10+
-   Windows / Linux

## Building

`git clone https://github.com/turtlecoin/go-karai`

Clone the repository

`go mod init github.com/turtlecoin/go-karai`

**First run only**: Initialize the go module

`GOPRIVATE='github.com/libp2p/*' go get ./...`

**First run only**: Look for available releases

`go build`

Compile to produce a binary `go-karai`

`go build -gcflags="-e" && ./go-karai`

**Optional:** Compile with all errors displayed, then run binary. Avoids "too many errors" from hiding error info.

## Contributing

-   `gofmt` is used on all files.
-   go modules are used to manage dependencies.
