#!/usr/bin/env python3
"""
nvchecker custom scraper for debuyoko-seiyuu-dicterm.
Fetches https://debuyoko.com/1388 and extracts the update date
from the pattern '※N件 YYYY/M/D', converting it to YYYYMMDD pkgver format.
"""
import re
import sys
import urllib.request

URL = 'https://debuyoko.com/1388'

try:
    with urllib.request.urlopen(URL, timeout=30) as resp:
        html = resp.read().decode('utf-8', errors='replace')
except Exception as e:
    print(f'ERROR: {e}', file=sys.stderr)
    sys.exit(1)

m = re.search(r'※\d+件\s+(\d{4})/(\d+)/(\d+)', html)
if not m:
    print('ERROR: date pattern not found', file=sys.stderr)
    sys.exit(1)

print(f'{m[1]}{int(m[2]):02d}{int(m[3]):02d}')
