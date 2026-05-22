#!/usr/bin/env python3
"""
nvchecker custom scraper for okaguchi-lawyer-dicterm.
Fetches https://togilab.com/download/dictermfile/ and extracts dateModified
from the JSON-LD schema, converting it to YYYYMMDD pkgver format.
"""
import json
import re
import sys
import urllib.request

URL = 'https://togilab.com/download/dictermfile/'

try:
    with urllib.request.urlopen(URL, timeout=30) as resp:
        html = resp.read().decode('utf-8', errors='replace')
except Exception as e:
    print(f'ERROR: {e}', file=sys.stderr)
    sys.exit(1)

# Extract dateModified from JSON-LD
m = re.search(r'<script[^>]+type=["\']application/ld\+json["\'][^>]*>(.*?)</script>',
              html, re.DOTALL)
if m:
    try:
        data = json.loads(m.group(1))
        date_str = data.get('dateModified', '')
        dm = re.match(r'(\d{4})-(\d{2})-(\d{2})', date_str)
        if dm:
            print(f'{dm[1]}{dm[2]}{dm[3]}')
            sys.exit(0)
    except (json.JSONDecodeError, KeyError):
        pass

# Fallback: visible text pattern "YYYY.MM.DD" near the download widget
m2 = re.search(r'法律用語辞書.*?(\d{4})\.(\d{2})\.(\d{2})', html, re.DOTALL)
if m2:
    print(f'{m2[1]}{m2[2]}{m2[3]}')
    sys.exit(0)

print('ERROR: dateModified not found', file=sys.stderr)
sys.exit(1)
