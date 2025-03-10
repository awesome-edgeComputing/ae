# AE (Advanced Edge) CLI Tool

AE is a powerful multi-language command-line tool designed for managing edge computing infrastructure, Kubernetes clusters, and system operations.

## Features

- **Multi-language Implementation**: Built with Golang, Python, and Rust
- **Modular Command Structure**: 
  - `esk`: Kubernetes cluster management (Golang)
  - `ei2`: Edge AI inference infrastructure management (Python)
  - `sys`: System utilities and monitoring (Rust)
- **Cross-platform Support**: Works on macOS, Linux (including ARM), and Windows
- **Single Binary Distribution**: All components are packaged into a single executable

## Installation

```bash
# Clone the repository
git clone https://github.com/your-org/ae.git
cd ae

# Install dependencies and build
make install
make build
```

## Usage

### ESK (Edge Service for Kubernetes)
```bash
ae esk list pod        # List all pods
ae esk get deployment  # Get deployment information
```

### EI2 (Edge Inference Infrastructure)
```bash
ae ei2 model list      # List available models
ae ei2 model deploy    # Deploy a model
ae ei2 infer start     # Start inference service
```

### SYS (System Utilities)
```bash
ae sys info           # Show system information
ae sys logs          # View system logs
ae sys monitor       # Monitor system resources
```

## Development

### Project Structure
```
ae/
├── cmd/
│   ├── esk/    # Golang implementation
│   ├── ei2/    # Python implementation
│   └── sys/    # Rust implementation
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   └── install.sh
└── Makefile
```

### Building from Source

```bash
# Build for all platforms
make build-all

# Build for specific platform
make build-darwin-amd64
make build-linux-amd64
make build-windows-amd64
```

## License

MIT License
