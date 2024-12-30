import os
from flask import Flask, request, jsonify
import requests

app = Flask(__name__)

DEEPSEEK_API_KEY = os.getenv('DEEPSEEK_API_KEY')
DEEPSEEK_API_URL = 'https://api.deepseek.com/v1/chat/completions'

@app.route('/')
def index():
    return 'Welcome to the DeepSeek API integration!'

@app.route('/chat', methods=['POST'])
def chat():
    data = request.json
    headers = {
        'Authorization': f'Bearer {DEEPSEEK_API_KEY}',
        'Content-Type': 'application/json'
    }
    response = requests.post(DEEPSEEK_API_URL, headers=headers, json=data)
    return jsonify(response.json())

if __name__ == '__main__':
    app.run(port=5000)