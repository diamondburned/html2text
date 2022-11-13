# html2text

[![Documentation](https://godoc.org/github.com/diamondburned/html2text?status.svg)](https://godoc.org/github.com/diamondburned/html2text)

html2text converts HTML to plain text.

This is useful for sending fancy HTML emails with an equivalently formatted text
as a fallback (e.g. for people who don't allow HTML emails or have other display
issues).

This package has been forked from the original html2text which removes some of
its quirks, such as having unnecessary punctuations or not supporting CJK.

## Install

```sh
go get github.com/diamondburned/html2text@latest
```

For usage, see the Go package documentation for an example.
