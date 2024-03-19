import requests

url = "http://192.168.122.1:10000/sony/camera"
def camera_req(method, params = []):
    json = {
    "method": method,
    "params": params,
    "id": 1,
    "version": "1.0"
    }
    req = requests.post(url, json = json)
    return req.status_code, req.json()

def get_method_types():
    _, res = camera_req("getMethodTypes")
    return res['results']

def get_liveview_status():
    _, res = camera_req("getEvent", [False])
    return res

def start_liveview():
    _, res = camera_req("stopLiveview", [])
    return res

c = get_liveview_status()
# print(c)

c = start_liveview()
print(c)

# http://192.168.122.1:60152/liveviewstream?%211234%21%2a%3a%2a%3aimage%2fjpeg%3a%2a%21%21%21%21%21