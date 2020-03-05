#!/usr/bin/env python3
import datetime
import json
import os
import requests
import time
from termcolor import colored
from pymongo import MongoClient


class PastebinCrawler:
    def get_timestamp(self):
        return time.strftime('%d/%m/%y %H:%M:%S')

    def log(self, message, type='neutral', clear_line=False):
        symbol = None
        if type == 'neutral':
            symbol = colored('[*]', 'yellow')
        elif type == 'negative':
            symbol = colored('[-]', 'red')
        elif type == 'positive':
            symbol = colored('[+]', 'green')
        elif type == 'warning':
            symbol = colored('[!]', 'magenta')
        else:
            raise Exception('Unknown log message type %s' % (type))

        date = colored(self.get_timestamp(), 'blue')

        line_ending = '\n'
        if clear_line:
            line_ending = '\r'
            rows, columns = os.popen('stty size', 'r').read().split()
            print(' '*int(columns), end='\r')

        print('%s %s %s' % (date, symbol, message), end=line_ending)

    def get_collection(self):
        client = MongoClient()
        db_name = self.get_db_name()
        db = client['pastebin']
        collection = db[db_name]
        client.close()
        return collection

    def get_db_name(self):
        today = datetime.datetime.today()
        year, month, day = today.isocalendar()
        return f'{year}-{month}'

    def save_to_db(self, paste):
        collection = self.get_collection()
        collection.insert_one(paste)

    def get_paste(self, key):
        url = 'https://scrape.pastebin.com/api_scrape_item.php?i=' + key
        retries = 10
        while retries > 0:
            r = requests.get(url)
            if r.status_code == 200:
                return r.text
            retries -= 1

    def db_contains(self, key):
        collection = self.get_collection()
        result = collection.find({'key': key})
        return result.count() > 0

    def make_request(self, url, convert_json=False, retries=10):
        while retries > 0:
            try:
                response = requests.get(url)
                if response.status_code == 200:
                    if convert_json:
                        return json.loads(response.text)
                    else:
                        return response.text
            except json.decoder.JSONDecodeError:
                pass
            retries -= 1
        raise Exception(f'Could not get {url}')

    def scrape(self):
        scrape_url = 'https://scrape.pastebin.com/api_scraping.php?limit=250'
        fields = ['date', 'key', 'expire', 'title', 'syntax', 'user',
                  'content']

        while True:
            data = self.make_request(scrape_url, convert_json=True)
            count = 0
            for paste in data:
                key = paste['key']
                paste_text = self.get_paste(key)
                paste['content'] = paste_text
                if not self.db_contains(key):
                    count += 1
                    filtered_paste = {}
                    for field in fields:
                        filtered_paste[field] = paste[field]
                    self.save_to_db(filtered_paste)
                    self.log(f'Saving {key}')
            if count:
                self.log(f'Found {count} new pastes', 'positive')
            self.log(f'Waiting 60 seconds...')
            time.sleep(60)


def main():
    crawler = PastebinCrawler()
    crawler.scrape()


if __name__ == '__main__':
    main()
