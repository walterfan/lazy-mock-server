"""Tests for the mock server."""

import pytest
import json
from unittest.mock import patch, mock_open
from app.mock_server import app


@pytest.fixture
def client():
    """Create a test client for the Flask app."""
    app.config['TESTING'] = True
    with app.test_client() as client:
        yield client


@pytest.fixture
def mock_config():
    """Mock YAML configuration."""
    return """
routes:
  - path: "/test"
    method: "GET"
    status_code: 200
    content_type: "application/json"
    response: {"message": "test response"}
  - path: "/text"
    method: "GET"
    status_code: 200
    content_type: "text/plain"
    response: "plain text response"
"""


def test_app_exists():
    """Test that the Flask app exists."""
    assert app is not None


def test_app_is_testing():
    """Test that the app is in testing mode."""
    app.config['TESTING'] = True
    assert app.config['TESTING'] is True


@patch("builtins.open", new_callable=mock_open)
@patch("os.path.exists")
def test_json_route(mock_exists, mock_file, client, mock_config):
    """Test JSON response route."""
    mock_exists.return_value = True
    mock_file.return_value.read.return_value = mock_config
    
    with patch("yaml.safe_load") as mock_yaml:
        mock_yaml.return_value = {
            'routes': [
                {
                    'path': '/test',
                    'method': 'GET',
                    'status_code': 200,
                    'content_type': 'application/json',
                    'response': {'message': 'test response'}
                }
            ]
        }
        
        response = client.get('/test')
        assert response.status_code == 200
        assert response.content_type == 'application/json'
        data = json.loads(response.data)
        assert data['message'] == 'test response'


@patch("builtins.open", new_callable=mock_open)
@patch("os.path.exists")
def test_text_route(mock_exists, mock_file, client, mock_config):
    """Test text/plain response route."""
    mock_exists.return_value = True
    mock_file.return_value.read.return_value = mock_config
    
    with patch("yaml.safe_load") as mock_yaml:
        mock_yaml.return_value = {
            'routes': [
                {
                    'path': '/text',
                    'method': 'GET',
                    'status_code': 200,
                    'content_type': 'text/plain',
                    'response': 'plain text response'
                }
            ]
        }
        
        response = client.get('/text')
        assert response.status_code == 200
        assert response.content_type == 'text/plain; charset=utf-8'
        assert response.data.decode() == 'plain text response'


def test_route_not_found(client):
    """Test 404 response for non-existent routes."""
    response = client.get('/nonexistent')
    assert response.status_code == 404
    data = json.loads(response.data)
    assert data['error'] == 'Route not found'


@patch("builtins.open", new_callable=mock_open)
@patch("os.path.exists")
def test_method_mismatch(mock_exists, mock_file, client, mock_config):
    """Test method mismatch returns 404."""
    mock_exists.return_value = True
    mock_file.return_value.read.return_value = mock_config
    
    with patch("yaml.safe_load") as mock_yaml:
        mock_yaml.return_value = {
            'routes': [
                {
                    'path': '/test',
                    'method': 'GET',
                    'status_code': 200,
                    'content_type': 'application/json',
                    'response': {'message': 'test response'}
                }
            ]
        }
        
        # Try POST instead of GET
        response = client.post('/test')
        assert response.status_code == 404
