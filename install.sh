REPO_URL="https://github.com/pimvandenbroek/podctl.git"
REPO_NAME=$(basename "$REPO_URL" .git)

# Clone the repository
echo "Cloning repository $REPO_URL..."
git clone "$REPO_URL"
cd "$REPO_NAME"

# Check for Go installation
if ! command -v go &> /dev/null; then
  echo "Go could not be found. Please install Go first."
  exit 1
fi

# Initialize Go modules if not already initialized (in case of go.mod missing)
if [ ! -f "go.mod" ]; then
  echo "No go.mod file found, initializing Go modules..."
  go mod init "$REPO_NAME"
  go mod tidy
else
  echo "Go modules found, installing dependencies..."
  go mod tidy
fi

# Build the Go application
echo "Building the application..."
go build $REPO_NAME.go

# Optionally, install the binary globally (comment this line if not needed)
echo "Installing the application globally..."
#go install

# Clean up: Delete the cloned repository
echo "Cleaning up the repository..."
mv $REPO_NAME ../$REPO_NAME.tmp
cd ..
rm -rf "$REPO_NAME"
mv $REPO_NAME.tmp $REPO_NAME

# Output success message
echo "Go application built and installed successfully!"

echo "To run the application, use: ./$REPO_NAME"