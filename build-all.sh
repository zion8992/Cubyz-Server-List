mkdir -p build

echo "Building for linux..."
GOOS=linux GOARCH=amd64 go build -o ironite-linux-amd64 ./src/

echo "Building for windows..."
GOOS=windows GOARCH=amd64 go build -o ironite-windows-amd64 ./src/

echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o ironite-macos-amd64 ./src/

echo "Moving..."
mv ironite-linux-amd64 build/
mv ironite-windows-amd64 build/
mv ironite-macos-amd64 build/

echo "Done"
