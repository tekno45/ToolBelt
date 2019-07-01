import json
import boto3
import os
from botocore.exceptions import ClientError

# Get keys from the bucket that match prefix
def get_photo_keys(prefix, bucket_obj):
    bucket = bucket_obj
    try:
        objects = bucket.objects.filter(Prefix=prefix)

        #count = objects
        for obj in objects:
            yield obj.key
        #print("total match key count: {}".format(objects))

    except ClientError as e:
        print("Error getting bucket objects")
    
# delete keys from s3 bucket
def delete_photo_keys(key, bucket, s3):
    print("deleted", key, "from", bucket.name)
    s3.Object(bucket.name, key).delete()
    return key

# invalidate objects in Cloud front
def create_cf_invalidation(keys, dist_id, request_id):
    print('starting invalidation')
    cf =boto3.client('cloudfront')
    result = cf.create_invalidation(DistributionId=dist_id, 
                                    InvalidationBatch={'Paths':
                                    {'Quantity': len(keys), 
                                    'Items':['/{}'.format(k) for k in keys]},
                                    'CallerReference': 'reference.{}'.format(request_id) })
    print(result)
    return result

        
def lambda_handler(event, context):
    ##Get variables as needed
    prefix = event['prefix']
    dist_id = os.environ.get('DIST_ID')
    s3= boto3.resource('s3')
    bucket=s3.Bucket(os.environ.get('BUCKET'))  
    keys=[]
    for pair in get_photo_keys(prefix=prefix, bucket_obj=bucket):
        delete_photo_keys(pair,bucket, s3)
        keys.append(pair)
    
    
    create_cf_invalidation(keys,dist_id, request_id=context.aws_request_id)
    return {
        'statusCode': 200,
        'body': json.dumps(keys)
        
    }
