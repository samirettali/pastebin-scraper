#!/usr/bin/env python3
import pymongo
import time
import utils
import os
import logging
from concurrent.futures import ThreadPoolExecutor


class PastebinCrawler:
    def __init__(self, db_address, db_name, col_name):
        self.set_logger()

        self.session = utils.get_retry_session()
        try:
            self.collection = utils.get_collection(db_address, db_name,
                                                   col_name)
        except pymongo.errors.ServerSelectionTimeoutError:
            self.logger.error('Cannot connect to database')

    def set_logger(self):
        log_dir = os.environ['LOG_DIR']
        formatter = logging.Formatter('%(asctime)s - %(levelname)s - '
                                      '%(message)s',
                                      datefmt='%d-%m-%y %H:%M:%S')

        fh = logging.FileHandler(f'{log_dir}/pastebin.log')
        fh.setFormatter(formatter)
        fh.setLevel(logging.INFO)

        self.logger = logging.getLogger(__name__)
        self.logger.addHandler(fh)

    def get_data(self):
        url = 'https://scrape.pastebin.com/api_scraping.php?limit=250'
        response = self.session.get(url)
        if response.status_code != 200:
            self.logger.error(f'{response.status_code} - {response.text}')
            exit(1)
        if 'Please wait a few minutes' in response.text:
            utils.log('API limit reached, waiting 10 minutes')
            time.sleep(600)
            exit(1)
        self.logger.info(f'{response.status_code} - {response.text}')
        return response.json()

    def check_paste(self, paste):
        key = paste['key']
        if not self.collection.find_one({'key': key}):
            paste_text = self.get_paste(key)
            del paste['scrape_url']
            del paste['full_url']
            del paste['size']
            paste['content'] = paste_text
            insertion = self.collection.insert_one(paste)
            assert insertion.inserted_id is not None
            return True
        return False

    def get_paste(self, key):
        url = f'https://scrape.pastebin.com/api_scrape_item.php?i={key}'
        response = self.session.get(url)
        error_msg = 'Error, we cannot find this paste.'

        if response.status_code == 200 and response.text != error_msg:
            return response.text
        return None

    def scrape(self):
        data = self.get_data()
        futures = []
        with ThreadPoolExecutor(max_workers=8) as executor:
            for paste in data:
                futures.append(executor.submit(self.check_paste, paste))

        count = 0
        for future in futures:
            if future.result():
                count += 1

        total = self.collection.count()
        self.logger.warning(f'{count:3} new pastes, {total} total')


def main():
    db_address = os.environ['DB_ADDRESS']
    db_name = os.environ['DB_NAME']
    col_name = os.environ['COL_NAME']
    healtcheck_url = os.environ['HEALTHCHECK']
    crawler = PastebinCrawler(db_address, db_name, col_name)
    crawler.scrape()
    utils.ping_healthchecks(healtcheck_url)


if __name__ == '__main__':
    main()
