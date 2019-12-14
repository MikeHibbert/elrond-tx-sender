import requests
import os


def get_proxies():
    url = "https://www.proxy-list.download/api/v1/get?type=http"

    r = requests.get(url)

    if r.status_code == 200:
        return r.text

    print(r.text)

    return None

if __name__ == "__main__":

    if os.path.exists('proxies.txt'):
        os.remove('proxies.txt')


    proxies = get_proxies()

    if proxies:
        with open('proxies.txt', "w") as proxies_file:
            proxies_file.write(proxies)
        
