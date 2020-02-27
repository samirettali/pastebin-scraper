#!/usr/bin/env python3
import json
import os
import re
import requests
import sqlite3
import time
from pathlib import Path
from termcolor import colored

connection = None
cursor = None
DB_FILENAME = 'scrape.db'
OUTPUT_DIRECTORY = 'loot'


class PastebinCrawler:

    def __init__(self):
        self.load_regexes()
        self.connect_db()

    # def table_exists(database, table):
    #     SELECT name FROM sqlite_master WHERE type='table' AND name='{table_name}';

    def connect_db(self):
        self.connection = sqlite3.connect(DB_FILENAME)
        self.cursor = self.connection.cursor()
        try:
            self.cursor.execute('SELECT * FROM scraped_bins')
        except sqlite3.OperationalError:
            self.log('Creating scraped_bins table', 'warning')
            self.cursor.execute('CREATE TABLE scraped_bins (key TEXT PRIMARY KEY, '
                        'interesting BOOLEAN NOT NULL, date TEXT NOT NULL)')
            self.connection.commit()


    def get_timestamp(self):
        return time.strftime('%d/%m/%y %H:%M:%S')


    def get_timestamp_as_path(self):
        timestamp = self.get_timestamp()
        timestamp = timestamp.replace('/', '_')
        timestamp = timestamp.replace(':', '_')
        timestamp = timestamp.replace(' ', '__')
        return timestamp


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


    def check_paste(self, key):
        url = 'https://scrape.pastebin.com/api_scrape_item.php?i=' + key
        paste_text = requests.get(url).text
        found = False
        for regex, file, directory in self.regexes:
            if re.match(regex, paste_text, re.IGNORECASE):
                log('Found a matching paste: %s -> %s' % (key, file), 'positive')
                self.save_result(paste_text, url, key, file, directory)
                found = True
        if not found:
            self.log(f'Paste {key} not matching', 'negative', True)
            timestamp = self.get_timestamp_as_path()
            path = f'{OUTPUT_DIRECTORY}/uncategorized/{timestamp}_{key}'
            f = open(path, 'w')
            f.write(paste_text)
            return 0
        return 1


    def save_result(self, paste_text, url, paste_id, file, directory):
        timestamp = self.get_timestamp()
        with open(f'{OUTPUT_DIRECTORY}/{file}', 'a') as matching:
            matching.write(f'{timestamp} - {url}\n')
        timestamp = self.get_timestamp_as_path()
        paste_id = paste_id.replace('/', '')
        path = f'{OUTPUT_DIRECTORY}/{directory}/{timestamp}_{paste_id}.txt'
        with open(path, 'w') as f:
            f.write(paste_text + '\n')


    def insert_in_db(self, key):
        timestamp = self.get_timestamp()
        interesting = self.check_paste(key)
        query_string = f'INSERT INTO scraped_bins (key, interesting, date) ' \
                    f'VALUES ("{key}", {interesting}, "{timestamp}")'
        self.connection.execute(query_string)
        self.connection.commit()


    def search_in_db(self, key):
        query_string = f'SELECT * FROM scraped_bins WHERE key = "{key}"'
        self.cursor.execute(query_string)
        return len(self.cursor.fetchall()) != 0


    def load_regexes(self):
        f = open('regexes.txt')
        self.regexes = []
        lines = f.read().splitlines()
        for line in lines:
            if len(line) > 0 and not line.startswith('#'):
                fields = line.split(',')
                self.regexes.append([field.strip() for field in fields])

        for i in range(len(self.regexes)):
            self.regexes[i] = [','.join(self.regexes[i][:-2])] + self.regexes[i][-2:]
            category = self.regexes[i][2]
            directory = f'{OUTPUT_DIRECTORY}/{category}'
            Path(directory).mkdir(parents=True, exist_ok=True)
        uncategorized_path = f'{OUTPUT_DIRECTORY}/uncategorized'
        Path(uncategorized_path).mkdir(parents=True, exist_ok=True)


    def scrape(self):
        SCRAPE_URL = 'https://scrape.pastebin.com/api_scraping.php?limit=250'

        while True:
            response = requests.get(SCRAPE_URL)
            data = json.loads(response.text)
            for paste in data:
                key = paste['key']
                if self.search_in_db(key):
                    log(f'Paste {key} already checked', 'negative', True)
                else:
                    self.insert_in_db(paste['key'])
            time.sleep(60)
    

def main():
    crawler = PastebinCrawler()
    crawler.scrape()


if __name__ == '__main__':
    main()
