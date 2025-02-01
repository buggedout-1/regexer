# Regexer

A simple Go application to process URLs and search for specific words in HTTP responses.

## Install
`git clone https://github.com/buggedout-1/regexer.git`  
`cd regexer`  
`go build regexer.go`  
`sudo cp regexer /usr/local/bin` 


## Usage
- `-u <url>`: Single URL to test.
- `-l <file_path>`: Path to a file with a list of URLs.
- `-w <word>`: Word to search for in the HTTP response body.

## Example:
```bash
go run regexer.go -u "https://example.com" -w "example"
go run regexer.go -l urls.txt -w "example"
```
or
```bash
regexer -u "https://example.com" -w "example"
regexer -l urls.txt -w "example"

