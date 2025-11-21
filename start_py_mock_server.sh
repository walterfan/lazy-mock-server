#!/bin/bash
#source ./venv/bin/activate

web_port=1989
web_debug=0

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <port=1989> <is_debug=0>"
else
  web_port="$1"
  web_debug="$2"
fi

export FLASK_DEBUG=${web_debug}
export FLASK_APP=app/mock_server.py

#flask run --host=0.0.0.0 --port=${web_port}
gunicorn  --pythonpath ./app --worker-class eventlet -b "0.0.0.0:${web_port}" -w 1 mock_server:app