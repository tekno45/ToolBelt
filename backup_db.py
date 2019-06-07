import gzip
import subprocess
import boto3
import argparse
import os
from botocore.exceptions import ClientError
import time

now = time.strftime("%y.%m.%d")
def backup_compress_db(database,backup_path, cmd):
    file_path= backup_path+database
    #open gz file
    with gzip.open('{0}.gz'.format(file_path), 'wb') as f:
        try:
            #start database dump to stdout
            popen = subprocess.Popen(cmd, stdout=subprocess.PIPE, universal_newlines=True)
        except Exception as e:
            print("Error running backup command", e)

        try:
            #direct stdout through gz
            for stdout_line in iter(popen.stdout.readline, ""):
                f.write(stdout_line.encode('utf-8'))
            #close file
            popen.stdout.close()
            popen.wait()
        except Exception as e:
                print("Error writing to backup location", e)
        return os.path.realpath(f.name)




if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-d", "--databases", nargs="+", required=True)
    parser.add_argument("--s3bucket")
    parser.add_argument("--s3key") #s3 directory path; 
    parser.add_argument("-b", "--backuppath")
    args = parser.parse_args()
    s3 = boto3.client("s3")

    backup_path=args.backuppath
    s3_bucket=args.s3bucket
    s3_path=args.s3key
    files = []
    # dump a database file for each name
    for database in args.databases:
        print(database, backup_path)
        cmd=["pg_dump", "{0}".format(database)]
        files.append(backup_compress_db(database,backup_path,cmd))
    # upload each dump file
    for f in files:
        with open(f,'rb') as y:
            path = "{1}/{0}.{2}".format(os.path.basename(y.name),s3_path,now)
            try:
              s3.upload_fileobj(y,s3_bucket,path)
              #s3.put_object(Bucket=s3_bucket,Key=path,Body=y)
              print("uploaded", path)
            except ClientError as e:
              print(e)
