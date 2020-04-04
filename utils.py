#!/usr/bin/env python3
import time
import requests
import pymongo
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry


def get_timestamp():
    return time.strftime('%d/%m/%y %H:%M:%S')


def get_retry_session(
    retries=3,
    backoff_factor=0.3,
    status_forcelist=(500, 502, 503, 504),
    session=None
):
    session = session or requests.Session()
    retry = Retry(
        total=retries,
        read=retries,
        connect=retries,
        backoff_factor=backoff_factor,
        status_forcelist=status_forcelist
    )
    adapter = HTTPAdapter(max_retries=retry)
    session.mount('http://', adapter)
    session.mount('https://', adapter)
    return session


def get_db(db_address, db_name):
    try:
        client = pymongo.MongoClient(db_address)
        client.server_info()
    except pymongo.errors.ServerSelectionTimeoutError:
        print(f'Could not connect to {db_address}')
        exit(1)
    db = client[db_name]
    client.close()
    return db


def get_collection(db_address, db_name, collection_name):
    db = get_db(db_address, db_name)
    collection = db[collection_name]
    return collection
