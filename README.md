# kustomize-gopass

kustomize-gopass is a CLI tool that generates Kubernetes secrets from gopass values.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Installation

To install kustomize-gopass, you need to have Go installed on your machine. Then, you can run:

```sh
go install jossware.com/kustomize-gopass@latest
```

## Usage

To use kustomize-gopass, run the following command:

```sh
kustomize-gopass
```

This will process your Kubernetes manifests and replace gopass placeholders with the actual secrets.

## Development

To build and run the project locally, clone the repository and run:

```sh
git clone https://github.com/yourusername/kustomize-gopass.git
cd kustomize-gopass
go build
./kustomize-gopass
```

## Running Tests

To run tests, use the following command:

```sh
go test ./...
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
