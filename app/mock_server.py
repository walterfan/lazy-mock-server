#!/usr/bin/env python3
import argparse
from flask import Flask, request, jsonify, Response
import yaml
import os

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))


app = Flask(__name__)

# Load the mock responses from the YAML file
def load_mock_responses(file_path):
    with open(file_path, 'r') as file:
        return yaml.safe_load(file)

mock_data = load_mock_responses(f'{CURRENT_DIR}/mock_response.yaml')

@app.route('/<path:path>', methods=['GET', 'POST'])
def mock_endpoint(path):
    global mock_data
    print(f"Request received: {request.method} {path}")
    # Find the matching route configuration
    for route in mock_data['routes']:
        if route['path'].strip('/') in path and route['method'].upper() == request.method:
            # Generate the response
            response_body = route['response']
            status_code = route.get('status_code', 200)
            content_type = route.get('content_type', 'application/json')
            
            if isinstance(response_body, dict):
                # Inject request data if specified
                if '{data}' in str(response_body):
                    response_body = {
                        k: (v.replace("{data}", str(request.json)) if isinstance(v, str) else v)
                        for k, v in response_body.items()
                    }
            
            # Return response based on content type
            if content_type == 'text/plain':
                return Response(str(response_body), status=status_code, mimetype='text/plain')
            elif content_type == 'application/json':
                return jsonify(response_body), status_code
            else:
                # For other content types, return as string with specified content type
                return Response(str(response_body), status=status_code, mimetype=content_type)

    # Default response for unmatched routes
    return jsonify({"error": "Route not found"}), 404

if __name__ == '__main__':
    # Parse command-line arguments for the port
    parser = argparse.ArgumentParser(description='Run a mock server.')
    parser.add_argument('--port', '-p', type=int, default=5000, help='Port to listen on (default: 5000)')
    args = parser.parse_args()

    # Start the Flask app on the specified port
    app.run(host='0.0.0.0', port=args.port, debug=True)