#!/bin/bash

echo "🔍 FinanCLI Setup Verification"
echo "=============================="

# Check Go version
echo "📋 Checking Go version..."
if command -v go &> /dev/null; then
    go_version=$(go version | awk '{print $3}')
    echo "✅ Go installed: $go_version"
    
    # Check if Go version is 1.21+
    version_check=$(go version | grep -E "go1\.(2[1-9]|[3-9][0-9])")
    if [ -n "$version_check" ]; then
        echo "✅ Go version is 1.21+"
    else
        echo "⚠️  Go version should be 1.21 or higher"
    fi
else
    echo "❌ Go not installed"
    echo "   Install from: https://golang.org/dl/"
fi

echo ""

# Check build
echo "🔨 Testing build..."
if make build > /dev/null 2>&1; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    echo "   Run 'make build' for details"
fi

echo ""

# Check tests
echo "🧪 Running tests..."
if make test > /dev/null 2>&1; then
    echo "✅ Tests passed"
else
    echo "❌ Tests failed"
    echo "   Run 'make test' for details"
fi

echo ""

# Check MongoDB (optional)
echo "🗄️  Checking MongoDB (optional)..."
if command -v mongod &> /dev/null; then
    echo "✅ MongoDB installed"
    
    # Check if MongoDB is running
    if pgrep -x "mongod" > /dev/null; then
        echo "✅ MongoDB is running"
    else
        echo "⚠️  MongoDB not running"
        echo "   Start with: brew services start mongodb/brew/mongodb-community"
    fi
else
    echo "⚠️  MongoDB not installed (optional)"
    echo "   Install with: brew install mongodb/brew/mongodb-community"
fi

echo ""

# Demo run
echo "🎮 Running demo..."
if make demo > /dev/null 2>&1; then
    echo "✅ Demo runs successfully"
else
    echo "❌ Demo failed"
    echo "   Run 'make demo' for details"
fi

echo ""
echo "🎯 Setup verification complete!"
echo ""
echo "📖 Available commands:"
echo "   make help     - Show all available commands"
echo "   make demo     - Run functionality demonstration"
echo "   make build    - Build the application"
echo "   make test     - Run test suite"
echo ""
echo "🚀 To run the full TUI application:"
echo "   make run      (requires interactive terminal)"