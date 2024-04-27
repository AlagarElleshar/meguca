#! /bin/env python3

import requests

while True:
    res = requests.get("https://git.io/GeoLite2-City.mmdb")
    if res.status_code == 404:
        m -= 1
        if m == 0:
            y -= 1
            m = 12
        continue
    res.raise_for_status()
    with open("GeoLite2-City.mmdb", "wb") as f:
        f.write(res.content)
        print("done")
        break
