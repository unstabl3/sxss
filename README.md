## Overview
This project is designed to enhance URL processing for security analysis by introducing concurrency and advanced error handling, including retry logic. It builds upon the ideas and initial work of Tom Hudson's `kxss` tool and extends its functionality to efficiently and effectively handle a large volume of URLs. Also I wrote this as a part of my learning golang. :)

## Features
- **Concurrency**: Utilize Go's concurrency model to process multiple URLs in parallel, significantly speeding up the analysis process.
- **Retry Logic**: Implements retry logic to handle transient errors, ensuring that temporary issues do not impede the scanning process.
- **Command-Line Arguments**: Offers flexibility in specifying the concurrency level and the number of retries, allowing users to balance performance and resource usage according to their system's capabilities.

## Installation

To install and run this tool, you'll need Go installed on your system. If you haven't already, follow the [official Go installation guide](https://golang.org/doc/install).

With Go installed, follow these steps:

```bash
go install github.com/unstabl3/sxss@latest
```

## Usage

To use the tool, pipe a list of URLs into the executable from the command line. You can also specify the concurrency level and the number of retries with flags:

```sh
cat urls.txt | sxss -concurrency 150 -retries 3
```

### Flags
- `-concurrency`: Number of concurrent goroutines for processing (default: 10)
- `-retries`: Maximum number of retries for each request (default: 3)

## Credits

This project is inspired by and builds upon the concepts introduced in [kxss](https://github.com/tomnomnom/hacks/blob/master/kxss/main.go) by Tom Hudson (@tomnomnom). Tom's original work on `kxss` provided valuable insights into processing URLs for XSS vulnerability scanning. All credit goes to him. :3


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


