#!/usr/bin/env python3
import argparse
import datetime
import json
import os
import pymongo
import requests
import time
import utils
from concurrent.futures import ThreadPoolExecutor
from termcolor import colored


class PastebinCrawler:
    def __init__(self, ip, port):
        self.db_address = f'mongodb://{ip}:{port}'

    def make_request(self, url, convert_json=False, retries=10):
        while retries > 0:
            try:
                response = requests.get(url)
                if response.status_code == 200:
                    if 'Please wait a few minutes' in response.text:
                        utils.log('API limit reached, waiting 10 minutes')
                        time.sleep(600)
                    if convert_json:
                        return json.loads(response.text)
                    else:
                        return response.text
            except json.decoder.JSONDecodeError:
                pass
            retries -= 1
        raise Exception(f'Could not get {url}')

    def get_db(self):
        try:
            client = pymongo.MongoClient(self.db_address)
            client.server_info()
        except pymongo.errors.ServerSelectionTimeoutError:
            print(f'Could not connect to {self.db_address}')
            exit(1)
        db = client['pastebin']
        client.close()
        return db

    def get_collection(self):
        db = self.get_db()
        collection_name = self.get_collection_name()
        collection = db[collection_name]
        return collection

    def get_collection_name(self):
        today = datetime.datetime.today()
        year, month, day = today.isocalendar()
        return f'{year}-{month}'

    def save_to_db(self, paste):
        self.collection.insert_one(paste)

    def db_contains(self, key):
        result = self.collection.find({'key': key})
        return result.count() > 0

    def get_db_size(self):
        db = self.get_db()
        stats = db.command('dbstats')
        size = stats['storageSize'] / 1048576
        return int(size)

    def check_paste(self, paste):
        fields = ['date', 'key', 'expire', 'title', 'syntax', 'user']
        key = paste['key']
        if not self.db_contains(key):
            paste_text = self.get_paste(key)
            filtered_paste = {}
            filtered_paste['content'] = paste_text
            for field in fields:
                filtered_paste[field] = paste[field]
            self.save_to_db(filtered_paste)
            return True
        return False

    def get_paste(self, key):
        url = 'https://scrape.pastebin.com/api_scrape_item.php?i=' + key
        retries = 10
        while retries > 0:
            r = requests.get(url)
            if r.status_code == 200:
                if 'Error, we cannot find this paste' in r.text:
                    return None
                else:
                    return r.text
            retries -= 1
        return None

    def scrape(self):
        scrape_url = 'https://scrape.pastebin.com/api_scraping.php?limit=250'
        while True:
            self.collection = self.get_collection()
            data = self.make_request(scrape_url, convert_json=True)
            futures = []
            with ThreadPoolExecutor(max_workers=8) as executor:
                for paste in data:
                    futures.append(executor.submit(self.check_paste, paste))
            count = 0
            for future in futures:
                if future.result():
                    count += 1

            if count:
                size = self.get_db_size()
                utils.log(f'{count:3} new pastes, collection size: {size} MB',
                          'positive')

            time.sleep(20)


def get_args():
    parser = argparse.ArgumentParser(description='Pastebin crawler')
    parser.add_argument('--ip', '-i', dest='ip', help='Mongodb ip address',
                        default='127.0.0.1')
    parser.add_argument('--port', '-p', dest='port', help='Mongodb port',
                        default='27017')
    args = parser.parse_args()
    return args.ip, args.port


def main():
    ip, port = get_args()
    crawler = PastebinCrawler(ip, port)
    crawler.scrape()


if __name__ == '__main__':
    main()
