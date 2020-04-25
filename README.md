![karai_github_banner](https://user-images.githubusercontent.com/34389545/80034381-f6a14d00-84b3-11ea-857a-638322dac890.png)

[![Discord](https://img.shields.io/discord/388915017187328002?label=Join%20Discord)](http://chat.turtlecoin.lol) [![GitHub issues](https://img.shields.io/github/issues/turtlecoin/go-karai?label=Issues)](https://github.com/turtlecoin/go-karai/issues) ![GitHub stars](https://img.shields.io/github/stars/turtlecoin/go-karai?label=Github%20Stars) ![Build](https://github.com/turtlecoin/go-karai/workflows/Build/badge.svg) ![GitHub](https://img.shields.io/github/license/turtlecoin/go-karai) ![GitHub issues by-label](https://img.shields.io/github/issues/turtlecoin/go-karai/Todo)

Read: [WHITEPAPER.md](https://github.com/turtlecoin/go-karai/blob/master/docs/WHITEPAPER.md)

Explore: [Karai Pointer Explorer](https://karaiexplorer.extrahash.org/)

## Usage

`./go-karai`

This will launch `go-karai`

Type `menu` to view a list of functions. Functions that are darkened are disabled.

## Dependencies

-   Golang 1.13+ [[Download]](https://golang.org)
-   TurtleCoin Daemon & Wallet-API [[Download]](http://latest.turtlecoin.lol)
-   Windows / Linux / MacOS

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

## Contributing Projects

[![turtlecoin](https://user-images.githubusercontent.com/34389545/80266529-fb0b6880-8661-11ea-9a75-4cb066834775.png)](https://turtlecoin.lol)
[![IPFS](https://user-images.githubusercontent.com/34389545/80266356-0c07aa00-8661-11ea-8308-84639318213a.png)](https://ipfs.io)
[![LibP2P](https://user-images.githubusercontent.com/34389545/80266502-e4651180-8661-11ea-8367-54bf59e26470.png)](https://libp2p.io)
[![GOLANG](https://user-images.githubusercontent.com/34389545/80266422-6b65ba00-8661-11ea-836a-d1904ec15b94.png)](https://golang.org)
