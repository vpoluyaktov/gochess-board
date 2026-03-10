#!/bin/bash

# Setup script for GoChess Board tests

echo "🧪 Setting up GoChess Board Tests..."

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install Node.js and npm first."
    echo "   Visit: https://nodejs.org/"
    exit 1
fi

# Install dependencies
echo "📦 Installing test dependencies..."
npm install

echo ""
echo "✅ Test setup complete!"
echo ""
echo "📝 To run tests:"
echo "   Browser:      npm run test:browser"
echo "   Command line: npm test"
echo "   Watch mode:   npm run test:watch"
echo ""
echo "📚 See README.md for more information"
