#!/bin/bash

# Lazy Mock Server - Python (Poetry) Startup Script
# Usage: ./start_py_mock_server_poetry.sh [port] [debug]

web_port=7777
web_debug=0

if [ "$#" -eq 1 ]; then
    web_port="$1"
elif [ "$#" -eq 2 ]; then
    web_port="$1"
    web_debug="$2"
elif [ "$#" -gt 2 ]; then
    echo "Usage: $0 [port=1989] [debug=0]"
    echo "Examples:"
    echo "  $0                    # Start on port 1989, debug off"
    echo "  $0 8080               # Start on port 8080, debug off"
    echo "  $0 8080 1             # Start on port 8080, debug on"
    exit 1
fi

# Check if Poetry is installed
if ! command -v poetry &> /dev/null; then
    echo "❌ Poetry is not installed. Please install Poetry first:"
    echo "   curl -sSL https://install.python-poetry.org | python3 -"
    echo "   Or visit: https://python-poetry.org/docs/#installation"
    exit 1
fi

# Check if pyproject.toml exists
if [ ! -f "pyproject.toml" ]; then
    echo "❌ pyproject.toml not found. Please run this script from the project root directory."
    exit 1
fi

echo "🚀 Starting Lazy Mock Server (Python/Flask with Poetry)"
echo "📁 Project: $(pwd)"
echo "🌐 Port: ${web_port}"
echo "🐛 Debug: ${web_debug}"
echo ""

# Install dependencies if needed
echo "📦 Installing dependencies..."
poetry install

# Set environment variables
export FLASK_DEBUG=${web_debug}
export FLASK_APP=app/mock_server.py

echo "🔥 Starting server..."
echo "📊 Web UI will be available at: http://localhost:${web_port}"
echo "🛑 Press Ctrl+C to stop"
echo ""

# Start the server using Poetry
if [ "${web_debug}" -eq 1 ]; then
    echo "🐛 Running in debug mode with Flask development server..."
    poetry run flask run --host=0.0.0.0 --port=${web_port}
else
    echo "🚀 Running in production mode with Gunicorn..."
    poetry run gunicorn --pythonpath ./app --worker-class eventlet -b "0.0.0.0:${web_port}" -w 1 mock_server:app
fi
