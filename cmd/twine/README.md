# Twine CLI

The official command-line tool for the Twine web framework.

## Installation

```bash
go install github.com/cstone-io/twine/cmd/twine@latest
```

## Usage

### Initialize a New Project

Create a new Twine project with a single command:

```bash
twine init my-app
```

This will:
- Create a new project directory
- Generate a basic application structure
- Set up templates and static file directories
- Create configuration files (.env.example, .gitignore)
- Generate a README with next steps

### Options

#### `--module` or `-m`
Specify the Go module path (default: `example.com/<project-name>`):

```bash
twine init my-app --module github.com/myuser/my-app
```

#### `--port` or `-p`
Set the server port (default: `3000`):

```bash
twine init my-app --port 8080
```

#### `--no-examples`
Skip generating example pages for a minimal setup:

```bash
twine init my-app --no-examples
```

#### `--with-db`
Include database setup (coming soon):

```bash
twine init my-app --with-db
```

#### `--with-auth`
Include authentication setup (coming soon):

```bash
twine init my-app --with-auth
```

### Other Commands

#### `version`
Show the CLI version:

```bash
twine version
```

## Generated Project Structure

```
my-app/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── .env.example              # Environment variables template
├── .gitignore                # Git ignore patterns
├── README.md                 # Project documentation
├── templates/
│   ├── pages/                # Full page templates
│   │   ├── index.html        # Home page
│   │   └── about.html        # About page
│   └── components/           # Reusable components
│       └── button.html       # Alpine Ajax button example
└── public/
    └── assets/               # Static files directory
        └── .gitkeep
```

## Quick Start

After creating a project:

```bash
# Navigate to project directory
cd my-app

# Run the application
go run main.go

# Visit in browser
http://localhost:3000
```

## Development

### Building from Source

```bash
git clone https://github.com/cstone-io/twine
cd twine/cmd/twine
go build -o twine
```

### Running Tests

```bash
go test ./...
```

## Future Commands

Coming soon:

- `twine dev` - Hot-reload development server
- `twine generate handler <name>` - Generate a handler file
- `twine generate model <name>` - Generate a model with migration
- `twine generate middleware <name>` - Generate custom middleware

## License

MIT
