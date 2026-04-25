import os
import yaml

files = ['docker-compose.yml', 'docker-compose.demo.yml']

def patch_compose(filepath):
    if not os.path.exists(filepath):
        print(f"File not found: {filepath}")
        return

    with open(filepath, 'r') as f:
        data = yaml.safe_load(f)

    # 1. Provide ml-service and remove old MLs
    ml_services = ['premium-ml', 'fraud-ml', 'forecast-ml']
    has_ml = False
    for ml in ml_services:
        if ml in data['services']:
            del data['services'][ml]
            has_ml = True
            
    if has_ml:
        data['services']['ml-service'] = {
            'build': {
                'context': './ml',
                'dockerfile': 'Dockerfile'
            },
            'ports': ['9000:8000'],
            'environment': {
                'SERVICE_PORT': 8000
            },
            'healthcheck': {
                'test': ["CMD-SHELL", "python -c \"import urllib.request; urllib.request.urlopen('http://localhost:8000/health', timeout=3)\""],
                'interval': '10s',
                'timeout': '5s',
                'retries': 15,
                'start_period': '20s'
            },
            'deploy': {
                'resources': {
                    'limits': {
                        'cpus': '0.5',
                        'memory': '512M'
                    }
                }
            }
        }

    # Remove zookeeper
    if 'zookeeper' in data['services']:
        del data['services']['zookeeper']
        
    for name, service in data['services'].items():
        # Remove zookeeper dependency everywhere
        if 'depends_on' in service:
            if isinstance(service['depends_on'], list):
                if 'zookeeper' in service['depends_on']:
                    service['depends_on'].remove('zookeeper')
            elif isinstance(service['depends_on'], dict):
                if 'zookeeper' in service['depends_on']:
                    del service['depends_on']['zookeeper']
            # Re-map premium-ml etc depending on ml-service
            for dt in ml_services:
                if isinstance(service['depends_on'], list):
                    if dt in service['depends_on']:
                        service['depends_on'].remove(dt)
                        if 'ml-service' not in service['depends_on']:
                            service['depends_on'].append('ml-service')
                elif isinstance(service['depends_on'], dict):
                    if dt in service['depends_on']:
                        cond = service['depends_on'][dt]
                        del service['depends_on'][dt]
                        if 'ml-service' not in service['depends_on']:
                            service['depends_on']['ml-service'] = cond

        # Update ML URLs in environment vars
        if 'environment' in service:
            if isinstance(service['environment'], dict):
                for key in ['ML_SERVICE_URL', 'PREMIUM_ML_URL', 'FRAUD_SERVICE_URL', 'FORECAST_ML_URL']:
                    if key in service['environment']:
                        service['environment'][key] = 'http://ml-service:8000'
                if any(
                    key in service['environment']
                    for key in ['ML_SERVICE_URL', 'PREMIUM_ML_URL', 'FRAUD_SERVICE_URL', 'FORECAST_ML_URL']
                ):
                    service['environment']['ML_TIMEOUT_MS'] = 3000
                    service['environment']['ML_RETRY_COUNT'] = 2
            elif isinstance(service['environment'], list):
                # not dict
                for i, env in enumerate(service['environment']):
                    if env.startswith('ML_SERVICE_URL=') or env.startswith('PREMIUM_ML_URL=') or env.startswith('FRAUD_SERVICE_URL=') or env.startswith('FORECAST_ML_URL='):
                        kw = env.split('=')[0]
                        service['environment'][i] = f"{kw}=http://ml-service:8000"

    # Configure Kafka for KRaft and Limits
    if 'kafka' in data['services']:
        kafka = data['services']['kafka']
        kafka['environment'] = {
            'KAFKA_NODE_ID': 1,
            'KAFKA_PROCESS_ROLES': 'broker,controller',
            'KAFKA_LISTENERS': 'PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093',
            'KAFKA_ADVERTISED_LISTENERS': 'PLAINTEXT://kafka:9092',
            'KAFKA_CONTROLLER_LISTENER_NAMES': 'CONTROLLER',
            'KAFKA_LISTENER_SECURITY_PROTOCOL_MAP': 'CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT',
            'KAFKA_CONTROLLER_QUORUM_VOTERS': '1@kafka:9093',
            'KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR': 1,
            'KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR': 1,
            'KAFKA_TRANSACTION_STATE_LOG_MIN_ISR': 1,
            'CLUSTER_ID': 'MkU3OEVBNTcwNTJENDM2Qk'
        }
        if 'deploy' not in kafka:
            kafka['deploy'] = {'resources': {'limits': {'cpus': '0.5', 'memory': '512M'}}}
            
    # Add limits to Postgres
    if 'postgres' in data['services']:
        if 'deploy' not in data['services']['postgres']:
            data['services']['postgres']['deploy'] = {'resources': {'limits': {'cpus': '1.0', 'memory': '1G'}}}

    # Add limits to Core
    if 'core' in data['services']:
        if 'deploy' not in data['services']['core']:
            data['services']['core']['deploy'] = {'resources': {'limits': {'cpus': '0.5', 'memory': '512M'}}}

    with open(filepath, 'w') as f:
        yaml.safe_dump(data, f, sort_keys=False)

for file in files:
    patch_compose(file)
print("done")
