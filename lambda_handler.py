import json
import boto3
from botocore.exceptions import ClientError

s3 = boto3.client('s3')

def lambda_handler(event, context):
    operation = event['operation']

    if operation == 'list':
        return list_buckets()
    elif operation == 'get':
        return get_object(event['bucket'], event['key'], event.get('presigned', False))
    elif operation == 'put':
        return put_object(event['bucket'], event['key'], event.get('body', None), event.get('presigned', False))
    elif operation == 'delete':
        return delete_object(event['bucket'], event['key'])
    else:
        return {
            'statusCode': 400,
            'body': json.dumps('Invalid operation')
        }

def list_buckets():
    response = s3.list_buckets()
    return {
        'statusCode': 200,
        'body': json.dumps(response['Buckets'])
    }

def get_object(bucket, key, presigned):
    if presigned:
        try:
            url = s3.generate_presigned_url(
                ClientMethod='get_object',
                Params={'Bucket': bucket, 'Key': key},
                ExpiresIn=3600
            )
            return {
                'statusCode': 200,
                'body': json.dumps({'url': url})
            }
        except ClientError as e:
            return {
                'statusCode': 400,
                'body': json.dumps(str(e))
            }
    else:
        response = s3.get_object(Bucket=bucket, Key=key)
        return {
            'statusCode': 200,
            'body': response['Body'].read().decode()
        }

def put_object(bucket, key, body=None, presigned=False):
    if presigned:
        try:
            url = s3.generate_presigned_url(
                ClientMethod='put_object',
                Params={'Bucket': bucket, 'Key': key},
                ExpiresIn=3600
            )
            return {
                'statusCode': 200,
                'body': json.dumps({'url': url})
            }
        except ClientError as e:
            return {
                'statusCode': 400,
                'body': json.dumps(str(e))
            }
    else:
        s3.put_object(Bucket=bucket, Key=key, Body=body)
        return {
            'statusCode': 200,
            'body': json.dumps('Object created')
        }

def delete_object(bucket, key):
    s3.delete_object(Bucket=bucket, Key=key)
    return {
        'statusCode': 200,
        'body': json.dumps('Object deleted')
    }
