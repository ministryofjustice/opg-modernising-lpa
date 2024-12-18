import json
import requests # type: ignore

def lambda_handler(event, context):
    response = requests.get('https://google.com')
    return {
        'statusCode': response.status_code,
        'body': response.text
    }


if __name__ == '__main__':
    output = lambda_handler("event", "contenxt")
    print(output)
