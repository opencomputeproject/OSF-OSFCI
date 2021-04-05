# Contributing to osfci

We need help with this project. Contributions are very welcome. See open [issues](https://github.com/HewlettPackard/osfci/issues).

## Bugs and Ideas

Please submit to https://github.com/HewlettPackard/osfci/issues

## Discussions and get in touch with us!

Please get in touch via https://github.com/HewlettPackard/osfci/discussions

## Coding Style

The ``osfci`` project aims to follow the standard formatting recommendations
and language idioms set out in the [Effective Go](https://golang.org/doc/effective_go.html)
guide, for example [formatting](https://golang.org/doc/effective_go.html#formatting)
and [names](https://golang.org/doc/effective_go.html#names).

`gofmt` and `golint` are required and will be checked during code-review.

- Example:
    ```
    gofmt -s -w ctrl1.go
    golint ctrl1.go
    ```

## Pull Requests

We accept GitHub pull requests.

Fork the project on GitHub, work in your fork and in branches, push
these to your GitHub fork, and when ready, do a GitHub pull requests
against https://github.com/HewlettPackard/osfci.

Every commit in your pull request needs to be able to build and pass our CI tests.

## Code Reviews

Look at the area of code you're modifying, its history, and consider
tagging some of the [maintainers](https://github.com/HewlettPackard/osfci/graphs/contributors) when doing a
pull request in order to instigate some code review.

## Quality Controls

Testing by the Advanced Technology Team and contributors.

## References

This document is inspired by standards followed by https://github.com/u-root/u-root