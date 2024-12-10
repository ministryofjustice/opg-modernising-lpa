import json
import requests

def lambda_handler(event, context):
    response = requests.get('https://google.com')
    return {
        'statusCode': 200,
        'body': json.dumps(response)
    }
