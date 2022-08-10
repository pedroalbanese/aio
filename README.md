# AIO
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://github.com/pedroalbanese/gzip/blob/master/LICENSE.md) 
[![GoDoc](https://godoc.org/github.com/pedroalbanese/gzip?status.png)](http://godoc.org/github.com/pedroalbanese/gzip)
[![Go Report Card](https://goreportcard.com/badge/github.com/pedroalbanese/gzip)](https://goreportcard.com/report/github.com/pedroalbanese/gzip)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pedroalbanese/gzip)](https://golang.org)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/pedroalbanese/gzip)](https://github.com/pedroalbanese/gzip/releases)
### All-in-One Command-line Compression Tool for modern multi-core machines written in Go 
<pre>Usage: aio [OPTION]... [FILE]
Compress or uncompress FILE (by default, compress FILE in-place).

 -a string
       compression algorithm: bzip2, lzma, xz, zlib, zstd (default "gzip")
 -c    write on standard output, keep original files unchanged
 -cores int
       number of cores to use for parallelization (default 1)
 -d    decompress; see also -c and -k
 -f    force overwrite of output file
 -h    print this help message
 -k    keep original files unchanged
 -s string
       use provided suffix on compressed files (default "gz")

With no FILE, or when FILE is -, read standard input.</pre>

## License

This project is licensed under the ISC License.

##### Copyright (c) 2020-2022 Pedro F. Albanese - ALBANESE Research Lab.
