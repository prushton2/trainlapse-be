import requests
import time
import json
import threading
import os
import sys
import asyncio
from http.server import ThreadingHTTPServer, BaseHTTPRequestHandler, HTTPServer

hostName = "0.0.0.0"
serverPort = 8080


def extractTripInfo(trip):
    return {
        "bikes_allowed": trip["attributes"]["bikes_allowed"],
        "block_id": trip["attributes"]["block_id"],
        "direction_id": trip["attributes"]["direction_id"],
        "headsign": trip["attributes"]["headsign"],
        "name": trip["attributes"]["name"],
        "revenue": trip["attributes"]["revenue"],
        "wheelchair_accessible": trip["attributes"]["wheelchair_accessible"]
    }

def bindTripInfo(vehicle, trip):
    del vehicle["relationships"]["trip"]
    vehicle["relationships"]["trip"] = trip

    return vehicle

def cleanVehicle(vehicle):

    vehicle["bearing"] = vehicle["attributes"]["bearing"]
    vehicle["label"] = vehicle["attributes"]["label"]
    vehicle["latitude"] = vehicle["attributes"]["latitude"]
    vehicle["longitude"] = vehicle["attributes"]["longitude"]
    vehicle["speed"] = vehicle["attributes"]["speed"]

    vehicle["headsign"] = vehicle["relationships"]["trip"]["headsign"]
    vehicle["name"] = vehicle["relationships"]["trip"]["name"]

    del vehicle["attributes"]
    del vehicle["id"]
    del vehicle["type"]
    del vehicle["links"]
    del vehicle["relationships"]

    return vehicle


async def getVehicles():
    print("Get Vehicles Thread started")
    while(1):
        if(int(time.time())%60 == 0):
            now = int(time.time())
            res = requests.get("https://api-v3.mbta.com/vehicles?filter[route_type]=2&include=trip")

            if(res.json().get("included") == None):
                print(f"    Failed to get data: {res.status_code} - {res.text}")
                continue

            inc = {}
            obj = []
            for i in res.json()["included"]:
                inc[i["id"]] = extractTripInfo(i)

            for i in res.json()["data"]:
                obj.append(bindTripInfo(i, inc[i["relationships"]["trip"]["data"]["id"]]))
                obj[-1] = cleanVehicle(obj[-1])

            with open(f"./out/{now}.json", 'w') as f:
                f.write(json.dumps(obj))

            for f in os.listdir("./out"):
                if int(f.split(".")[0]) < now - 86400:
                    os.remove(os.path.join("./out", f))
            # print("    Saved data to file.")
        await asyncio.sleep(1)

class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type", "application/json")
        self.end_headers()

        if(self.path.split("/")[1].startswith("get")):
            query = self.path.split("?")[1]
            query = query.split("&")
            query = {i.split("=")[0]: i.split("=")[1] for i in query}

            jsonobj = {
                "size": 0,
                "elements": []
            }

            for f in sorted(os.listdir("./out")):
                if int(time.time()) - int(f.split(".")[0]) <= int(query["time"]):
                    with open(os.path.join("./out", f), 'r') as file:
                        jsonobj["size"] += 1
                        jsonobj["elements"].append(
                            {
                                "data": json.loads(file.read()),
                                "timestamp": int(f.split(".")[0])
                            }
                        )

            self.wfile.write(bytes(json.dumps(jsonobj), "utf-8"))


if __name__ == "__main__":
    print("Starting server...")
    webServer = ThreadingHTTPServer((hostName, serverPort), MyServer)
    print("Server started http://%s:%s" % (hostName, serverPort))

    t1 = threading.Thread(target=lambda: asyncio.run(getVehicles()))

    t1.start()

    webServer.serve_forever()

