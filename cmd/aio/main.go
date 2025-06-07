// Copyright (c) 2010, Andrei Vieru. All rights reserved.
// Copyright (c) 2021, Pedro F. Albanese. All rights reserved.
// Copyright (c) 2025: Pindorama
//	Luiz Antônio Rangel (takusuman)
// All rights reserved.
// Use of this source code is governed by a ISC license that
// can be found in the LICENSE file.
package main

import (
	"compress/gzip"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/zstd"
	"github.com/pedroalbanese/brotli"
	"github.com/pedroalbanese/lzma"
	"github.com/pedroalbanese/xz"
	"rsc.io/getopt"
)

// Command-line flags
var (
	stdout     = flag.Bool("c", false, "write on standard output, keep original files unchanged")
	decompress = flag.Bool("d", false, "decompress; see also -c and -k")
	force      = flag.Bool("f", false, "force overwrite of output file")
	help       = flag.Bool("h", false, "print this help message")
	verbose    = flag.Bool("v", false, "be verbose")
	keep       = flag.Bool("k", false, "keep original files unchanged")
	suffix     = flag.String("S", "gz", "use provided suffix on compressed files")
	cores      = flag.Int("cores", 0, "number of cores to use for parallelization")
	test       = flag.Bool("t", false, "test compressed file integrity")
	level      = flag.Int("l", 4, "compression level (1 = fastest, 9 = best)")
	recursive  = flag.Bool("r", false, "operate recursively on directories")
	algorithm  = flag.String("algorithm", "gzip", "brotli, gzip, zlib, bzip2, s2, zstd, lzma, xz")

	stdin bool // Indicates if reading from standard input
)

// usage displays program usage instructions
func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [FILE] ...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Compress or uncompress FILEs (by default, compress FILEs in-place).\n\n")
	getopt.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nWith no FILE, or when FILE is -, read standard input.\n")
	fmt.Fprintf(os.Stderr, "\nSupported algorithms:\n")
	fmt.Fprintf(os.Stderr, "  brotli - Google's Brotli algorithm\n")
	fmt.Fprintf(os.Stderr, "  gzip   - GNU zip compression (default)\n")
	fmt.Fprintf(os.Stderr, "  zlib   - zlib compression\n")
	fmt.Fprintf(os.Stderr, "  bzip2  - bzip2 compression\n")
	fmt.Fprintf(os.Stderr, "  s2     - Snappy2 compression (fast)\n")
	fmt.Fprintf(os.Stderr, "  zstd   - Zstandard compression\n")
	fmt.Fprintf(os.Stderr, "  lzma   - LZMA compression\n")
	fmt.Fprintf(os.Stderr, "  xz     - XZ compression (LZMA2)\n")
}

// exit shows an error message and exits the program with error code
func exit(msg string) {
	usage()
	fmt.Fprintln(os.Stderr)
	log.Fatalf("%s: check args: %s\n\n", os.Args[0], msg)
}

// setByUser checks whether a specific flag was explicitly set by the user
func setByUser(name string) (isSet bool) {
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			isSet = true
		}
	})
	return
}

// getDefaultSuffix returns the default suffix for the current algorithm
func getDefaultSuffix() string {
	switch *algorithm {
	case "brotli":
		return "br"
	case "zlib":
		return "zlib"
	case "bzip2":
		return "bz2"
	case "s2":
		return "s2"
	case "zstd":
		return "zst"
	case "lzma":
		return "lzma"
	case "xz":
		return "xz"
	default:
		return "gz"
	}
}

// getAlgorithmFromExtension returns the compression algorithm based on the file extension
func getAlgorithmFromExtension(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "", fmt.Errorf("file has no extension")
	}

	// Remove the leading dot (e.g., ".gz" → "gz")
	ext = ext[1:]

	switch ext {
	case "gz":
		return "gzip", nil
	case "zlib":
		return "zlib", nil
	case "bz2":
		return "bzip2", nil
	case "s2":
		return "s2", nil
	case "zst":
		return "zstd", nil
	case "lzma":
		return "lzma", nil
	case "xz":
		return "xz", nil
	case "br":
		return "brotli", nil
	default:
		return "", fmt.Errorf("unsupported extension: .%s", ext)
	}
}

// processFile processes a single file (compression, decompression, or test)
// Returns an error if any issue occurs during processing
func processFile(inFilePath string) error {
	if *decompress && inFilePath != "-" {
		detectedAlgo, err := getAlgorithmFromExtension(inFilePath)
		if err != nil {
			return fmt.Errorf("failed to detect algorithm: %v", err)
		}

		// Override algorithm if user did not set it explicitly
		if !setByUser("algorithm") {
			*algorithm = detectedAlgo
		}

		// Set suffix based on detected algorithm, if user did not set it
		if !setByUser("S") {
			*suffix = getDefaultSuffix()
		}
	} else {
		// Set default suffix if not provided by user
		if !setByUser("S") {
			*suffix = getDefaultSuffix()
		}
	}

	// Checks for conflicting flags
	if *stdout == true && setByUser("S") == true {
		return fmt.Errorf("stdout set, suffix not used")
	}
	if *stdout == true && *force == true {
		return fmt.Errorf("stdout set, force not used")
	}
	if *stdout == true && *keep == true {
		return fmt.Errorf("stdout set, keep is redundant")
	}

	var outFilePath string // Output file path

	// Test mode: verifies compressed file integrity
	if *test {
		var inFile *os.File
		var err error
		if inFilePath == "-" {
			inFile = os.Stdin
		} else {
			inFile, err = os.Open(inFilePath)
			if err != nil {
				return err
			}
			defer inFile.Close()
		}

		var r io.Reader
		switch *algorithm {
		case "gzip":
			gr, err := gzip.NewReader(inFile)
			if err != nil {
				return fmt.Errorf("test failed: %v", err)
			}
			defer gr.Close()
			r = gr
		case "zlib":
			zr, err := zlib.NewReader(inFile)
			if err != nil {
				return fmt.Errorf("test failed: %v", err)
			}
			defer zr.Close()
			r = zr
		case "bzip2":
			br, err := bzip2.NewReader(inFile, nil)
			if err != nil {
				return fmt.Errorf("test failed: %v", err)
			}
			r = br
		case "s2":
			r = s2.NewReader(inFile)
		case "zstd":
			zr, err := zstd.NewReader(inFile)
			if err != nil {
				return fmt.Errorf("test failed: %v", err)
			}
			defer zr.Close()
			r = zr
		case "lzma":
			lr := lzma.NewReader(inFile)
			r = lr
		case "xz":
			xr, err := xz.NewReader(inFile)
			if err != nil {
				return fmt.Errorf("test failed: %v", err)
			}
			r = xr
		default: // brotli
			r = brotli.NewReader(inFile)
		}

		_, err = io.Copy(io.Discard, r)
		if err != nil {
			return fmt.Errorf("test failed: %v", err)
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "%s: OK\n", inFilePath)
		}
		return nil
	}

	// Determines the input source (stdin or file)
	if inFilePath == "-" { // read from stdin
		if *stdout != true {
			return fmt.Errorf("reading from stdin, can write only to stdout")
		}
		if setByUser("S") == true {
			return fmt.Errorf("reading from stdin, suffix not needed")
		}
		stdin = true
	} else { // read from file
		f, err := os.Lstat(inFilePath)
		if err != nil {
			return err
		}
		if f == nil {
			return fmt.Errorf("file %s not found", inFilePath)
		}
		if f.IsDir() {
			return fmt.Errorf("%s is a directory", inFilePath)
		}

		// Determines the output destination (file)
		if !*stdout { // write to file
			if *suffix == "" {
				return fmt.Errorf("suffix can't be an empty string")
			}

			// Generates output file name
			if *decompress {
				outFileDir, outFileName := path.Split(inFilePath)
				if strings.HasSuffix(outFileName, "."+*suffix) {
					if len(outFileName) > len("."+*suffix) {
						nstr := strings.SplitN(outFileName, ".", len(outFileName))
						estr := strings.Join(nstr[0:len(nstr)-1], ".")
						outFilePath = outFileDir + estr
					} else {
						return fmt.Errorf("can't strip suffix .%s from file %s", *suffix, inFilePath)
					}
				} else {
					return fmt.Errorf("file %s doesn't have suffix .%s", inFilePath, *suffix)
				}
			} else {
				outFilePath = inFilePath + "." + *suffix
			}

			// Checks if output file already exists
			f, err = os.Lstat(outFilePath)
			if err == nil && f != nil {
				if !*force {
					return fmt.Errorf("outFile %s exists. use -f to overwrite", outFilePath)
				}
				if f.IsDir() {
					return fmt.Errorf("outFile %s is a directory", outFilePath)
				}
				err = os.Remove(outFilePath)
				if err != nil {
					return err
				}
			}
		}
	}

	// Creates a pipe for communication between goroutines
	pr, pw := io.Pipe()

	// File decompression
	if *decompress {
		go func() {
			defer pw.Close()
			var inFile *os.File
			var err error
			if inFilePath == "-" {
				inFile = os.Stdin
			} else {
				inFile, err = os.Open(inFilePath)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
				defer inFile.Close()
			}

			if *verbose {
				fmt.Fprintf(os.Stderr, "%s: ", inFile.Name())
			}

			_, err = io.Copy(pw, inFile)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}()

		var r io.Reader
		var err error
		switch *algorithm {
		case "gzip":
			gr, err := gzip.NewReader(pr)
			if err != nil {
				pr.Close()
				return err
			}
			defer gr.Close()
			r = gr
		case "zlib":
			zr, err := zlib.NewReader(pr)
			if err != nil {
				pr.Close()
				return err
			}
			defer zr.Close()
			r = zr
		case "bzip2":
			r, err = bzip2.NewReader(pr, nil)
			if err != nil {
				return fmt.Errorf("corrupted file or format error: %v", err)
			}
		case "s2":
			r = s2.NewReader(pr)
		case "zstd":
			zr, err := zstd.NewReader(pr)
			if err != nil {
				pr.Close()
				return err
			}
			defer zr.Close()
			r = zr
		case "lzma":
			r = lzma.NewReader(pr)
		case "xz":
			xr, err := xz.NewReader(pr)
			if err != nil {
				pr.Close()
				return err
			}
			r = xr
		default: // brotli
			r = brotli.NewReader(pr)
		}

		var outFile *os.File
		if *stdout {
			outFile = os.Stdout
		} else {
			outFile, err = os.Create(outFilePath)
			if err != nil {
				pr.Close()
				return err
			}
			defer outFile.Close()
		}

		_, err = io.Copy(outFile, r)
		pr.Close()
		if err != nil {
			return err
		}

		if *verbose && !*stdout {
			fmt.Fprintln(os.Stderr, "done")
		}
	} else { // File compression
		go func() {
			defer pw.Close()
			var inFile *os.File
			var err error
			if inFilePath == "-" {
				inFile = os.Stdin
			} else {
				inFile, err = os.Open(inFilePath)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
				defer inFile.Close()
			}

			// Cria um contador para a saída
			counter := &writeCounter{Writer: pw}

			var w io.WriteCloser
			switch *algorithm {
			case "gzip":
				w, err = gzip.NewWriterLevel(counter, *level)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
			case "zlib":
				w, err = zlib.NewWriterLevel(counter, *level)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
			case "bzip2":
				w, err = bzip2.NewWriter(counter, &bzip2.WriterConfig{Level: *level})
				if err != nil {
					pw.CloseWithError(err)
					return
				}
			case "s2":
				switch {
				case *level <= 3:
					w = s2.NewWriter(counter, s2.WriterBetterCompression())
				case *level >= 7:
					w = s2.NewWriter(counter, s2.WriterBestCompression())
				default:
					w = s2.NewWriter(counter)
				}
			case "zstd":
				w, err = zstd.NewWriter(counter, zstd.WithEncoderLevel(zstd.EncoderLevel(*level)))
				if err != nil {
					pw.CloseWithError(err)
					return
				}
			case "lzma":
				w = lzma.NewWriterLevel(counter, *level)
			case "xz":
				w, err = xz.NewWriter(counter)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
			default: // brotli
				w = brotli.NewWriterLevel(counter, *level)
			}

			if *verbose {
				fmt.Fprintf(os.Stderr, "%s: ", inFile.Name())

				// Conta os bytes de entrada
				inSize, err := io.Copy(w, inFile)
				if err != nil {
					w.Close()
					pw.CloseWithError(err)
					return
				}

				if err := w.Close(); err != nil {
					pw.CloseWithError(err)
					return
				}

				// Calcula as estatísticas
				outSize := counter.bytesWritten
				var ratio float64
				if outSize > 0 {
					ratio = float64(inSize) / float64(outSize)
				} else {
					ratio = 0
				}

				fmt.Fprintf(os.Stderr, "%6.3f:1, %6.3f bits/byte, %5.2f%% saved, %d in, %d out.\n",
					ratio,
					(8 / ratio), // bits por byte
					(100 * (1 - (1 / ratio))),
					inSize,
					outSize)
			} else {
				// Versão sem verbose
				_, err = io.Copy(w, inFile)
				if err != nil {
					w.Close()
					pw.CloseWithError(err)
					return
				}
				if err := w.Close(); err != nil {
					pw.CloseWithError(err)
					return
				}
			}
		}()

		var outFile *os.File
		var err error
		if *stdout {
			outFile = os.Stdout
		} else {
			outFile, err = os.Create(outFilePath)
			if err != nil {
				pr.Close()
				return err
			}
			defer outFile.Close()
		}

		_, err = io.Copy(outFile, pr)
		pr.Close()
		if err != nil {
			return err
		}
	}

	// Removes the original file if needed
	if !*stdout && !*keep && inFilePath != "-" {
		err := os.Remove(inFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

type writeCounter struct {
	io.Writer
	bytesWritten int64
}

func (w *writeCounter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.bytesWritten += int64(n)
	return n, err
}

// main is the program's entry point
func main() {
	// Configure flags for compression levels (1–9)
	for i := 1; i <= 9; i++ {
		explanation := fmt.Sprintf("compression level %d", i)
		if i == 4 {
			explanation += " (default)"
		}
		_ = flag.Bool(strconv.Itoa(i), false, explanation)
	}

	// Alias short flags with their long counterparts.
	getopt.Aliases(
		"1", "fast",
		"9", "best",
		"c", "stdout",
		"d", "decompress",
		"f", "force",
		"k", "keep",
		"r", "recursive",
		"t", "test",
		"v", "verbose",
		"h", "help",
	)

	// Parse command-line flags
	getopt.Parse()

	// Check if someone has used '-#' for a compression level.
	if !setByUser("l") {
		for i := 1; i <= 11; i++ {
			if setByUser(strconv.Itoa(i)) {
				*level = i
				break
			}
		}
	}

	// Validate compression level
	if *level < 1 || *level > 11 {
		exit("invalid compression level: must be between 1 and 9")
	}

	// Validate algorithm
	validAlgorithms := map[string]bool{
		"brotli": true,
		"gzip":   true,
		"zlib":   true,
		"bzip2":  true,
		"s2":     true,
		"zstd":   true,
		"lzma":   true,
		"xz":     true,
	}
	if !validAlgorithms[strings.ToLower(*algorithm)] {
		exit(fmt.Sprintf("invalid algorithm: %s", *algorithm))
	}

	// Show help if requested
	if *help {
		usage()
		os.Exit(0)
	}

	// Validate number of cores
	if setByUser("cores") && (*cores < 1 || *cores > 32) {
		exit("invalid number of cores")
	}

	// From 'go doc runtime.GOMAXPROCS':
	// "It defaults to the value of runtime.NumCPU.
	// If n < 1, it does not change the current setting."
	// In fact, if the default value of cores is zero, it
	// will use all the cores of the machine.
	runtime.GOMAXPROCS(*cores)

	// Get list of files to process
	files := flag.Args()
	if len(files) == 0 {
		files = []string{"-"} // default to stdin
	}

	// Process each file
	hasErrors := false
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			log.Printf("%s: %v", file, err)
			hasErrors = true
			continue
		}

		if info.IsDir() {
			if *recursive {
				err = filepath.Walk(file, func(path string, fi os.FileInfo, err error) error {
					if err != nil {
						log.Printf("%s: %v", path, err)
						hasErrors = true
						return nil
					}
					if !fi.IsDir() {
						if err := processFile(path); err != nil {
							log.Printf("%s: %v", path, err)
							hasErrors = true
						}
					}
					return nil
				})
				if err != nil {
					log.Printf("%s: %v", file, err)
					hasErrors = true
				}
			} else {
				log.Printf("%s is a directory (use -r to process recursively)", file)
				hasErrors = true
			}
		} else {
			err := processFile(file)
			if err != nil {
				log.Printf("%s: %v", file, err)
				hasErrors = true
			}
		}
	}

	// Exit with error code if any failures occurred
	if hasErrors {
		os.Exit(1)
	}
}
