import requests
import time
import json
import threading
import os
import sys
import asyncio
from http.server import BaseHTTPRequestHandler, HTTPServer

hostName = "localhost"
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
    print("getVehicles")
    while(1):
        if(int(time.time())%10 == 0):
            now = int(time.time())
            res = requests.get("https://api-v3.mbta.com/vehicles?filter[route_type]=2&include=trip")

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
                if int(f.split(".")[0]) < now - 600:
                    os.remove(os.path.join("./out", f))

        await asyncio.sleep(1)

class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):
        print(self.path.split("/"))
        if(self.path.split("/")[1].startswith("get")):
            query = self.path.split("?")[1]
            query = query.split("&")
            query = {i.split("=")[0]: i.split("=")[1] for i in query}
            print(query)

            self.wfile.write(bytes("{", "utf-8"))

            for f in os.listdir("./out"):
                if time.time() - int(f.split(".")[0]) <= int(query["time"]):
                    with open(os.path.join("./out", f), 'r') as file:
                        # print(file.read())
                        self.wfile.write(bytes(f"\"{int(time.time())}\": {file.read()}", "utf-8"))

            self.wfile.write(bytes("}", "utf-8"))

        self.send_response(200)
        self.send_header("Content-type", "text/json")
        # self.end_headers()
        # self.wfile.write(bytes("<html><head><title>https://pythonbasics.org</title></head>", "utf-8"))
        # self.wfile.write(bytes("<p>Request: %s</p>" % self.path, "utf-8"))
        # self.wfile.write(bytes("<body>", "utf-8"))
        # self.wfile.write(bytes("<p>This is an example web server.</p>", "utf-8"))
        # self.wfile.write(bytes("</body></html>", "utf-8"))

if __name__ == "__main__":        
    webServer = HTTPServer((hostName, serverPort), MyServer)
    print("Server started http://%s:%s" % (hostName, serverPort))

    t = threading.Thread(target=lambda: asyncio.run(getVehicles()))
    t.start()

    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass
    
    print("Server stopped.")

    webServer.server_close()
    print("Server stopped.")


