#!/usr/bin/env python3
import json
import requests
import sqlite3
import os
import time
import re
from pathlib import Path
from termcolor import colored
from sys import argv

regexes = []
connection = None
cursor = None
DB_FILENAME = 'scrape.db'
OUTPUT_DIRECTORY = 'loot'


def connect_db():
    global connection
    global cursor
    connection = sqlite3.connect(DB_FILENAME)
    cursor = connection.cursor()
    try:
        cursor.execute('SELECT * FROM scraped_bins')
    except sqlite3.OperationalError:
        log('Creating scraped_bins table', 'warning')
        cursor.execute('CREATE TABLE scraped_bins (key TEXT PRIMARY KEY, '
                       'interesting BOOLEAN NOT NULL, date text not null)')
        connection.commit()


def get_timestamp():
    return time.strftime('%d/%m/%y %H:%M:%S')


def get_timestamp_as_path():
    timestamp = get_timestamp()
    timestamp = timestamp.replace('/','_')
    timestamp = timestamp.replace(':','_')
    timestamp = timestamp.replace(' ','__')
    return timestamp


def log(message, type='neutral', clear_line = False):
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

    date = colored(get_timestamp(), 'blue')

    line_ending = '\n'
    if clear_line:
        line_ending = '\r'
        rows, columns = os.popen('stty size', 'r').read().split()
        print(' '*int(columns), end='\r')

    print('%s %s %s' % (date, symbol, message), end=line_ending)


def check_paste(key):
    global regexes
    url = 'https://scrape.pastebin.com/api_scrape_item.php?i=' + key
    paste_text = requests.get(url).text
    found = False
    for regex, file, directory in regexes:
        if re.match(regex, paste_text, re.IGNORECASE):
            log('Found a matching paste: %s -> %s' % (key, file), 'positive')
            save_result(paste_text, url, key, file, directory)
            found = True
    if not found:
        log(f'Paste {key} not matching', 'negative', True)
        timestamp = get_timestamp_as_path()
        path = f'{OUTPUT_DIRECTORY}/uncategorized/{timestamp}_{key}'
        f = open(path, 'w')
        f.write(paste_text)
        return 0
    return 1


def save_result(paste_text, url, paste_id, file, directory):
    timestamp = get_timestamp()
    with open(f'{OUTPUT_DIRECTORY}/{file}', 'a') as matching:
        matching.write(f'{timestamp} - {url}\n')
    # try:
        # os.mkdir(directory)
    # except KeyboardInterrupt:
        # raise
    # except:
        # pass
    timestamp = get_timestamp_as_path()
    paste_id = paste_id.replace('/', '')
    with open(f'{OUTPUT_DIRECTORY}/{directory}/{timestamp}_{paste_id}.txt', 'w') as f:
        f.write(paste_text + '\n')


def insert_in_db(key):
    timestamp = get_timestamp()
    interesting = check_paste(key)
    query_string = f'INSERT INTO scraped_bins (key, interesting, date) ' \
                   f'VALUES ("{key}", {interesting}, "{timestamp}")'
    connection.execute(query_string)
    connection.commit()


def search_in_db(key):
    query_string = f'SELECT * FROM scraped_bins WHERE key = "{key}"'
    cursor.execute(query_string)
    return len(cursor.fetchall()) != 0


def load_regexes():
    global regexes
    f = open('regexes.txt')
    lines = f.read().splitlines()
    for line in lines:
        if len(line) > 0 and not line.startswith('#'):
            fields = line.split(',')
            regexes.append([field.strip() for field in fields])

    for i in range(len(regexes)):
        regexes[i] = [','.join(regexes[i][:-2])] + regexes[i][-2:]
        category = regexes[i][2]
        directory = f'{OUTPUT_DIRECTORY}/{category}'
        Path(directory).mkdir(parents=True, exist_ok=True)
    Path(f'{OUTPUT_DIRECTORY}/uncategorized').mkdir(parents=True, exist_ok=True)


def scrape():
    global connection
    global cursor
    global regexes

    SCRAPING_URL = 'https://scrape.pastebin.com/api_scraping.php?limit=250'
    BIN_URL = 'https://scrape.pastebin.com/api_scrape_item_meta.php?i='

    while True:
        response = requests.get(SCRAPING_URL)
        data = json.loads(response.text)
        for paste in data:
            key = paste['key']
            if search_in_db(key):
                log(f'Paste {key} already checked', 'negative', True)
            else:
                insert_in_db(paste['key'])
        time.sleep(60)


def main():
    load_regexes()
    connect_db()
    scrape()


if __name__ == '__main__':
    main()
