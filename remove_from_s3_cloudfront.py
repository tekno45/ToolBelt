import gzip

import boto3
from botocore.exceptions import ClientError
from datetime import datetime
import argparse
import flask

# Get keys from the bucket that match prefix
def get_photo_keys(prefix, bucket_obj):
    bucket = bucket_obj

    objects = bucket.objects.filter(Prefix=prefix)
    for obj in objects:
        yield obj.key

# copy keys to backup bucket
def move_to_backup_bucket(key, bucket, dest_bucket):
    s3 = boto3.resource('s3')
    s3.Object(dest_bucket.name, '{0}.bak'.format(key)).copy_from(CopySource='{0}/{1}'.format(bucket.name,key))
    print('copied', key)
    old_obj = s3.Object(bucket.name, key)
    response = old_obj.delete()['DeleteMarker']
    return response

# invalidate objects in Cloud front
def create_cf_invalidation(prefix, dist_id):
    print('invalidation') 
    #remove wild card character for invalidation
    if prefix[::1] == "*":
        prefix = prefix[:-1]

    now = datetime.now()
    cf =boto3.client('cloudfront')
    result = cf.create_invalidation(DistributionId=dist_id, 
                                    InvalidationBatch={'Paths':
                                    {'Quantity': 1, 
                                    'Items':['/{}'.format(prefix)]},
                                    'CallerReference': 'reference.{}'.format(now) })
    return result

# main function
def delete_request(prefix, bucket, dist_id, dest_bucket):
    s3= boto3.resource('s3')
    bucket=s3.Bucket(bucket)
    dest_bucket= boto3.resource('s3').Bucket(dest_bucket)
    try:
        keys=[key for key in get_photo_keys(prefix=prefix, bucket_obj=bucket)]
    except ClientError as e:
        exit()

    for key in keys:
        move_to_backup_bucket(key,bucket, dest_bucket)
    

    
    if(len(keys)):
        create_cf_invalidation(prefix ,dist_id)
    else: print("No keys match") 
if __name__ == "__main__":
    ##Argument parsing
    parser = argparse.ArgumentParser()
    parser.add_argument('-p','--prefix', required=True, type=str)
    args = parser.parse_args()
    prefix= args.prefix

    ##Environment constants
    bucket='ng.movoto.com'
    dist_id='E1NGL15C45YBI2' #pi.movoto.com cloud front distribution
    dest_bucket='photo.backups.movoto.com'

    ##Main run
    delete_request(prefix, bucket, dist_id, dest_bucket) 
        
    
   
    
