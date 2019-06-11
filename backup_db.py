import gzip
import subprocess
import boto3
import os
from botocore.exceptions import ClientError
import time
import json

now = time.strftime("%y.%m.%d")
conf_path = "backuppy.conf"
def backup_compress_db(database,backup_path, cmd):
    """ Dump Database, compress it in memory, and return path to file"""
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

def read_conf():
    """ read config file and return json of options"""
    try:
        with open(conf_path, "r") as conf:
            config = json.loads(conf)
            return config["configuration"]
    except Exception as e:
        print("Cannot read from config file at: ", conf_path)
        exit()

if __name__ == "__main__":
    s3 = boto3.client("s3")
    config = read_conf()
    backup_path= config["backup_path"]
    s3_bucket=config["s3_bucket"]
    s3_path=config["s3_path"]
    databases = config["databases"]
    files = []
    # dump a database file for each name
    for database in databases:
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
