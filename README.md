# AIO Compress
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://github.com/pedroalbanese/aio/blob/master/LICENSE.md) 
[![GoDoc](https://godoc.org/github.com/pedroalbanese/aio?status.png)](http://godoc.org/github.com/pedroalbanese/aio)
[![GitHub downloads](https://img.shields.io/github/downloads/pedroalbanese/aio/total.svg?logo=github&logoColor=white)](https://github.com/pedroalbanese/aio/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pedroalbanese/aio)](https://goreportcard.com/report/github.com/pedroalbanese/aio)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pedroalbanese/aio)](https://golang.org)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/pedroalbanese/aio)](https://github.com/pedroalbanese/aio/releases)
### All-in-One Command-line Compression Tool for modern multi-core machines 
<pre>Usage: aio [OPTION]... [FILE]...
Compress or uncompress FILEs (by default, compress FILEs in-place).

  -1, --fast
        compression level 1
  -2    compression level 2
  -3    compression level 3
  -4    compression level 4 (default)
  -5    compression level 5
  -6    compression level 6
  -7    compression level 7
  -8    compression level 8
  -9, --best
        compression level 9 (4 for zstd and 11 for brotli)
  -S string
        use provided suffix on compressed files (default "gz")
  --algorithm string
        brotli, gzip, zlib, bzip2, s2, zstd, lzma, xz (default "gzip")
  -c, --stdout
        write on standard output, keep original files unchanged
  --cores int
        number of cores to use for parallelization
  -d, --decompress
        decompress; see also -c and -k
  -f, --force
        force overwrite of output file
  -h, --help
        print this help message
  -k, --keep
        keep original files unchanged
  -l int
        compression level (1 = fastest, 9 = best) (default 4)
  -r, --recursive
        operate recursively on directories
  -t, --test
        test compressed file integrity
  -v, --verbose
        be verbose

With no FILE, or when FILE is -, read standard input.</pre>

## License

This project is licensed under the ISC License.

##### Copyright (c) 2020-2025 Pedro F. Albanese - ALBANESE Research Lab.
