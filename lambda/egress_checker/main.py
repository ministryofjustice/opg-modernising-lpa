import json
import requests # type: ignore

def lambda_handler(event, context):
    response = requests.get('https://google.com')
    print(response)
    return {
        'statusCode': 200,
    }
