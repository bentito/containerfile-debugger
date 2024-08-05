
# Containerfile Debugger

## Overview

The Containerfile Debugger is a tool that helps you debug your Containerfiles by allowing you to set breakpoints and enter an interactive shell at any line. This can be particularly useful for debugging and inspecting the build process of your container images.

## Prerequisites

- [Podman](https://podman.io/getting-started/installation) installed on your system.
- [Go](https://golang.org/doc/install) installed on your system.

## Setup

1. Ensure you have a `Containerfile-debug-buildah` in the project directory with the following content:

    ```Dockerfile
    # Use Fedora as the base image
    FROM fedora:latest

    # Install Buildah and necessary tools
    RUN dnf -y install buildah fuse-overlayfs

    # Set up the environment for running Buildah
    ENV STORAGE_DRIVER=vfs

    # Set entrypoint to Buildah
    ENTRYPOINT ["buildah"]
    ```

2. Run the setup script to build the Podman image, create the `buildah-run.sh` script, and set up necessary directories:

    ```sh
    chmod +x setup-buildah.sh
    ./setup-buildah.sh
    ```

## Usage

### Set a Breakpoint

To set a breakpoint at a specific line in your Containerfile:

```sh
go run main.go set-breakpoint Containerfile <line-number>
```

### Build with Debugger

To build the Containerfile with the debugger, which will pause at each breakpoint and open an interactive shell:

```sh
go run main.go build Containerfile
```

### Continue Build

If needed, you can use the `continue` command to resume the build process from the last breakpoint:

```sh
go run main.go continue
```

## Example

### Example Containerfile

```Dockerfile
# Use an official Python runtime as a parent image
FROM python:3.10-slim-buster

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Install any needed packages
# BREAKPOINT
RUN pip install --no-cache-dir Flask

# Make port 80 available to the world outside this container
EXPOSE 80

# Define environment variable
ENV NAME World

# Add a simple Flask app
RUN echo "from flask import Flask\napp = Flask(__name__)\n@app.route('/')\ndef hello():\n    return 'Hello, ${NAME}!'\n\nif __name__ == '__main__':\n    app.run(host='0.0.0.0', port=80)" > app.py

# Run app.py when the container launches
CMD ["python", "app.py"]
```

### Setting a Breakpoint

To set a breakpoint at line 11:

```sh
go run main.go set-breakpoint Containerfile 11
```

### Building with Debugger

To build the Containerfile and enter an interactive shell at the breakpoint:

```sh
go run main.go build Containerfile
```

## Contributing

Feel free to submit issues, fork the repository, and send pull requests. Contributions are welcome!

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
