#!/usr/bin/env python3
import os
import time
from termcolor import colored

def get_timestamp():
    return time.strftime('%d/%m/%y %H:%M:%S')

def log(message, type='neutral', clear_line=False):
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

