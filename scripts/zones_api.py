import json
from flask import Flask, jsonify
import os

app = Flask(__name__)

MAX_ZONE_NAMES_PER_LEVEL = 15

@app.route('/api/zones/<level>', methods=['GET'])
def get_zones(level):
    base_dir = os.path.dirname(__file__)
    if level == 'a':
        path = os.path.join(base_dir, '../zone_a.json')
    elif level == 'b':
        path = os.path.join(base_dir, '../zone_b.json')
    elif level == 'c':
        path = os.path.join(base_dir, '../zone_c.json')
    else:
        return jsonify({'error': 'Invalid zone level'}), 400
    try:
        with open(path, encoding='utf-8') as f:
            data = json.load(f)
            return jsonify(data[:MAX_ZONE_NAMES_PER_LEVEL])
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5050, debug=True)
