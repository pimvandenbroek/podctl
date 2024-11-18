# podctl

`podctl` is a command-line tool designed to simplify executing in containers. With `podctl`, you can easily list and interact with your Kubernetes resources, and execute a shell inside a container with minimal effort.

## Features

- Execute a shell inside a container, by interactively going through cluster/namespace/pod

## Installation

To install `podctl`, you can use the following command:

```sh
curl -sSL https://github.com/pimvandenbroek/podctl/raw/main/install.sh | sh
```

#### Alternatively, you can build it yourself

#### Clone repo

```sh
git clone https://github.com/pimvandenbroek/podctl.git
cd podctl
```

#### Install dependencies and build

```sh
go mod tidy
go build
# optional
go install
```

## Usage

```sh
podctl
```

## Contributing

We welcome contributions to `podctl`. If you would like to contribute, please fork the repository and submit a pull request.

## License

`podctl` is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

## Contact

For any questions or feedback, please open an issue on the [GitHub repository](https://github.com/pimvandenbroek/podctl).
